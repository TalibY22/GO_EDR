<script setup>
import { ref, onMounted } from 'vue';


const FilterMatchMode = {
    EQUALS: 'equals',
    CONTAINS: 'contains',
    DATE_RANGE: 'dateRange'
};

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
    <div class="grid">
        <!-- Filters Card -->
        <div class="col-12">
            <div class="card">
                <div class="flex align-items-center gap-3">
                    <div class="w-12rem">
                        <Dropdown
                            v-model="filters.event.value"
                            :options="eventTypes"
                            placeholder="Event Type"
                            class="w-full"
                        />
                    </div>
                    <div class="w-12rem">
                        <InputText
                            v-model="filters.agent_id.value"
                            placeholder="Agent ID"
                            class="w-full"
                        />
                    </div>
                    <div class="w-12rem">
                        <InputText
                            v-model="filters.details.value"
                            placeholder="Search Details"
                            class="w-full"
                        />
                    </div>
                    <div class="w-20rem">
                        <Calendar
                            v-model="filters.date_range.value"
                            selectionMode="range"
                            dateFormat="yy-mm-dd"
                            showTime
                            placeholder="Select Date Range"
                            class="w-full"
                            showIcon
                        />
                    </div>
                    <div class="flex gap-2">
                        <Button 
                            icon="pi pi-search"
                            label="Apply Filters"
                            @click="onFilter"
                            severity="primary"
                            size="small"
                        />
                        <Button 
                            icon="pi pi-filter-slash" 
                            label="Clear" 
                            @click="clearFilters"
                            outlined
                            size="small"
                        />
                    </div>
                </div>
            </div>
        </div>

        <!-- Table Card -->
        <div class="col-12">
            <div class="card">
                <div class="flex justify-content-between align-items-center mb-4">
                    <h5 class="m-0">Log History</h5>
                    <Button 
                        icon="pi pi-download" 
                        label="Export CSV" 
                        @click="exportCSV"
                        severity="success"
                        size="small"
                    />
                </div>

                <DataTable
                    :value="logs"
                    :lazy="true"
                    :paginator="true"
                    :rows="10"
                    :totalRecords="totalRecords"
                    :loading="loading"
                    :rows-per-page-options="[10,25,50]"
                    paginator-template="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
                    current-page-report-template="Showing {first} to {last} of {totalRecords} entries"
                    response-lazy-load-event="page"
                    @page="onPage"
                    stripedRows
                    class="p-datatable-lg"
                >
                    <Column field="id" header="ID" sortable>
                        <template #body="{ data }">
                            <span class="font-bold">#{{ data.id }}</span>
                        </template>
                    </Column>

                    <Column field="event" header="Event" sortable>
                        <template #body="{ data }">
                            <span 
                                :class="getSeverityClass(data.event)"
                                class="px-2 py-1 rounded-full text-xs font-semibold"
                            >
                                {{ data.event }}
                            </span>
                        </template>
                    </Column>

                    <Column field="agent_id" header="Agent ID" sortable />

                    <Column field="details" header="Details" sortable>
                        <template #body="{ data }">
                            <div class="whitespace-pre-wrap">{{ data.details }}</div>
                        </template>
                    </Column>

                    <Column field="created_at" header="Created At" sortable>
                        <template #body="{ data }">
                            {{ new Date(data.created_at).toLocaleString() }}
                        </template>
                    </Column>

                    <Column :exportable="false" style="min-width: 8rem">
                        <template #body="{ data }">
                            <Button
                                icon="pi pi-copy"
                                text
                                rounded
                                @click="navigator.clipboard.writeText(data.details)"
                                v-tooltip.top="'Copy Details'"
                            />
                        </template>
                    </Column>
                </DataTable>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Base styles */
.card {
    background: var(--surface-card);
    padding: 1.5rem;
    border-radius: var(--border-radius);
    margin-bottom: 1rem;
}

.p-datatable {
    margin-top: 1rem;
}

:deep(.p-calendar) {
    width: 100%;
}

:deep(.p-dropdown),
:deep(.p-inputtext) {
    height: 2.5rem;
}

/* Make details column wider */
:deep(.p-datatable-wrapper) td:nth-child(4) {
    max-width: 400px;
}

/* Responsive adjustments */
@media screen and (max-width: 960px) {
    .flex.align-items-center.gap-3 {
        flex-wrap: wrap;
    }
    
    .w-12rem, .w-20rem {
        width: 100% !important;
        margin-bottom: 0.5rem;
    }
}
</style> 