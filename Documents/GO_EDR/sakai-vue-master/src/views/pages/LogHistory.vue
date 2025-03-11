<script setup>
import { ref, onMounted } from 'vue';
import { FilterMatchMode } from 'primevue/api';

const logs = ref([]);
const loading = ref(false);
const totalRecords = ref(0);
const lazyParams = ref({
    first: 0,
    rows: 10,
    page: 1,
    sortField: 'created_at',
    sortOrder: -1
});

// Filter states
const filters = ref({
    event: { value: null, matchMode: FilterMatchMode.EQUALS },
    agent_id: { value: null, matchMode: FilterMatchMode.EQUALS },
    details: { value: null, matchMode: FilterMatchMode.CONTAINS },
    date_range: { value: null, matchMode: FilterMatchMode.DATE_RANGE }
});

const eventTypes = ref([
    'ERROR',
    'WARNING',
    'INFO',
    'BASH',
    'COMMAND',
    'CONNECTION'
]);

// Fetch logs with pagination and filters
const loadLogs = async () => {
    loading.value = true;
    try {
        // Build query parameters
        const params = new URLSearchParams({
            page: lazyParams.value.page.toString(),
            limit: lazyParams.value.rows.toString()
        });

        // Add filters
        if (filters.value.event.value) {
            params.append('event', filters.value.event.value);
        }
        if (filters.value.agent_id.value) {
            params.append('agent_id', filters.value.agent_id.value);
        }
        if (filters.value.date_range.value) {
            const [startDate, endDate] = filters.value.date_range.value;
            if (startDate) params.append('start_date', startDate.toISOString());
            if (endDate) params.append('end_date', endDate.toISOString());
        }

        const response = await fetch(`http://localhost:8080/logs?${params.toString()}`);
        const data = await response.json();
        
        logs.value = data.data;
        totalRecords.value = data.pagination.total;
    } catch (error) {
        console.error('Error loading logs:', error);
    } finally {
        loading.value = false;
    }
};

// Handle page/sort changes
const onPage = (event) => {
    lazyParams.value = event;
    loadLogs();
};

// Handle filter changes
const onFilter = () => {
    lazyParams.value.page = 1;
    loadLogs();
};

// Clear all filters
const clearFilters = () => {
    filters.value = {
        event: { value: null, matchMode: FilterMatchMode.EQUALS },
        agent_id: { value: null, matchMode: FilterMatchMode.EQUALS },
        details: { value: null, matchMode: FilterMatchMode.CONTAINS },
        date_range: { value: null, matchMode: FilterMatchMode.DATE_RANGE }
    };
    onFilter();
};

// Export logs to CSV
const exportCSV = () => {
    const csvContent = [
        // CSV Header
        ['ID', 'Event', 'Agent ID', 'Details', 'Created At'],
        // CSV Data
        ...logs.value.map(log => [
            log.id,
            log.event,
            log.agent_id,
            log.details,
            new Date(log.created_at).toLocaleString()
        ])
    ].map(row => row.join(',')).join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `log_export_${new Date().toISOString()}.csv`;
    link.click();
};

// Get severity class for event type
const getSeverityClass = (event) => {
    switch (event) {
        case 'ERROR':
            return 'bg-red-100 text-red-800';
        case 'WARNING':
            return 'bg-yellow-100 text-yellow-800';
        case 'INFO':
            return 'bg-blue-100 text-blue-800';
        case 'BASH':
            return 'bg-purple-100 text-purple-800';
        default:
            return 'bg-gray-100 text-gray-800';
    }
};

onMounted(() => {
    loadLogs();
});
</script>

<template>
    <div class="card">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">Log History</h1>
            <div class="flex gap-2">
                <Button 
                    icon="pi pi-filter-slash" 
                    label="Clear Filters" 
                    @click="clearFilters"
                    class="p-button-outlined"
                />
                <Button 
                    icon="pi pi-download" 
                    label="Export CSV" 
                    @click="exportCSV"
                    class="p-button-success"
                />
            </div>
        </div>

        <DataTable
            :value="logs"
            :lazy="true"
            :paginator="true"
            :rows="10"
            :total-records="totalRecords"
            :loading="loading"
            :filters="filters"
            :rows-per-page-options="[10,25,50]"
            paginator-template="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
            current-page-report-template="Showing {first} to {last} of {totalRecords} entries"
            response-lazy-load-event="page"
            @page="onPage"
            @filter="onFilter"
            striped-rows
            class="p-datatable-lg"
        >
            <Column field="id" header="ID" sortable>
                <template #body="{ data }">
                    <span class="font-semibold">#{{ data.id }}</span>
                </template>
            </Column>

            <Column field="event" header="Event" sortable>
                <template #filter="{ filterModel }">
                    <Dropdown
                        v-model="filterModel.value"
                        :options="eventTypes"
                        placeholder="Select Event"
                        class="p-column-filter"
                        show-clear
                    />
                </template>
                <template #body="{ data }">
                    <span 
                        :class="getSeverityClass(data.event)"
                        class="px-2 py-1 rounded-full text-xs font-semibold"
                    >
                        {{ data.event }}
                    </span>
                </template>
            </Column>

            <Column field="agent_id" header="Agent ID" sortable>
                <template #filter="{ filterModel }">
                    <InputText
                        v-model="filterModel.value"
                        type="text"
                        class="p-column-filter"
                        placeholder="Search Agent ID"
                    />
                </template>
            </Column>

            <Column field="details" header="Details" sortable>
                <template #filter="{ filterModel }">
                    <InputText
                        v-model="filterModel.value"
                        type="text"
                        class="p-column-filter"
                        placeholder="Search Details"
                    />
                </template>
                <template #body="{ data }">
                    <div class="whitespace-pre-wrap">{{ data.details }}</div>
                </template>
            </Column>

            <Column field="created_at" header="Created At" sortable>
                <template #filter="{ filterModel }">
                    <Calendar
                        v-model="filterModel.value"
                        selection-mode="range"
                        date-format="yy-mm-dd"
                        placeholder="Date Range"
                        class="p-column-filter"
                        show-time
                    />
                </template>
                <template #body="{ data }">
                    {{ new Date(data.created_at).toLocaleString() }}
                </template>
            </Column>

            <Column :exportable="false" style="min-width: 8rem">
                <template #body="{ data }">
                    <Button
                        icon="pi pi-copy"
                        class="p-button-rounded p-button-text"
                        @click="navigator.clipboard.writeText(data.details)"
                        v-tooltip.top="'Copy Details'"
                    />
                </template>
            </Column>
        </DataTable>
    </div>
</template>

<style scoped>
.p-datatable {
    margin-top: 1rem;
}

:deep(.p-column-filter) {
    width: 100%;
}

/* Adjust calendar width in filter */
:deep(.p-calendar) {
    width: 100%;
}

/* Make details column wider */
:deep(.p-datatable-wrapper) td:nth-child(4) {
    max-width: 400px;
}
</style> 