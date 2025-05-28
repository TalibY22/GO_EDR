<script setup>
import { onMounted, ref } from 'vue';
import { useToast } from 'primevue/usetoast';
import { FilterMatchMode } from '@primevue/core/api';

const API_URL = 'http://localhost:8080';

onMounted(() => {
    loadAgents();
});

const toast = useToast();
const dt = ref();
const agents = ref([]);
const agentDialog = ref(false);
const deleteAgentDialog = ref(false);
const deleteAgentsDialog = ref(false);
const agent = ref({});
const selectedAgents = ref();
const filters = ref({
    global: { value: null, matchMode: FilterMatchMode.CONTAINS }
});
const submitted = ref(false);
const loading = ref(false);
const operatingSystems = ref([
    { label: 'Linux', value: 'linux' },
    { label: 'Windows', value: 'windows' },
    { label: 'macOS', value: 'macos' }
]);
const statuses = ref([
    { label: 'Online', value: 'online' },
    { label: 'Offline', value: 'offline' },
    { label: 'Unknown', value: 'unknown' }
]);

async function loadAgents() {
    loading.value = true;
    try {
        const response = await fetch(`${API_URL}/agents`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
        agents.value = data.data;
    } catch (error) {
        console.error('Error loading agents:', error);
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to load agents', life: 3000 });
    } finally {
        loading.value = false;
    }
}

function formatDate(dateString) {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    return date.toLocaleString();
}

function getStatusSeverity(status) {
    switch (status) {
        case 'online':
            return 'success';
        case 'offline':
            return 'danger';
        case 'unknown':
        default:
            return 'warn';
    }
}

function openNew() {
    agent.value = {
        name: '',
        os: null,
        status: 'unknown',
        description: ''
    };
    submitted.value = false;
    agentDialog.value = true;
}

function hideDialog() {
    agentDialog.value = false;
    submitted.value = false;
}

async function saveAgent() {
    submitted.value = true;

    if (agent.value.name?.trim() && agent.value.os) {
        try {
            // Format the agent data before sending
            const agentData = {
                name: agent.value.name,
                os: agent.value.os.value || agent.value.os,
                status: agent.value.status.value || agent.value.status,
                description: agent.value.description || ''
            };

            if (agent.value.id) {
                // Update agent - Note: Your API doesn't have an update endpoint yet
                // const response = await fetch(`${API_URL}/agents/${agent.value.id}`, {
                //     method: 'PUT',
                //     headers: { 'Content-Type': 'application/json' },
                //     body: JSON.stringify(agentData)
                // });
                // For now, show a message that updating isn't implemented
                toast.add({ severity: 'warn', summary: 'Warning', detail: 'Update API not implemented yet', life: 3000 });
            } else {
                // Create new agent
                const response = await fetch(`${API_URL}/agents`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(agentData)
                });
                
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                
                toast.add({ severity: 'success', summary: 'Successful', detail: 'Agent Created', life: 3000 });
                await loadAgents();
            }

            agentDialog.value = false;
            agent.value = {};
        } catch (error) {
            console.error('Error saving agent:', error);
            toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save agent', life: 3000 });
        }
    }
}

function editAgent(ag) {
    agent.value = { ...ag };
    agentDialog.value = true;
}

function confirmDeleteAgent(ag) {
    agent.value = ag;
    deleteAgentDialog.value = true;
}

async function deleteAgent() {
    try {
        const response = await fetch(`${API_URL}/agents/${agent.value.id}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        await loadAgents();
        deleteAgentDialog.value = false;
        agent.value = {};
        toast.add({ severity: 'success', summary: 'Successful', detail: 'Agent Deleted', life: 3000 });
    } catch (error) {
        console.error('Error deleting agent:', error);
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete agent', life: 3000 });
    }
}

function confirmDeleteSelected() {
    deleteAgentsDialog.value = true;
}

async function deleteSelectedAgents() {
    try {
        // Delete multiple agents in sequence
        for (const ag of selectedAgents.value) {
            const response = await fetch(`${API_URL}/agents/${ag.id}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
        }
        
        await loadAgents();
        deleteAgentsDialog.value = false;
        selectedAgents.value = null;
        toast.add({ severity: 'success', summary: 'Successful', detail: 'Agents Deleted', life: 3000 });
    } catch (error) {
        console.error('Error deleting agents:', error);
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete agents', life: 3000 });
    }
}

function downloadAgent(ag) {
    try {
        window.location.href = `${API_URL}/agents/${ag.id}/download`;
        toast.add({ severity: 'info', summary: 'Download', detail: 'Agent download started', life: 3000 });
    } catch (error) {
        console.error('Error downloading agent:', error);
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to download agent', life: 3000 });
    }
}

function exportCSV() {
    dt.value.exportCSV();
}
</script>

<template>
    <div>
        <div class="card">
            <Toolbar class="mb-6">
                <template #start>
                    <Button label="New" icon="pi pi-plus" severity="secondary" class="mr-2" @click="openNew" />
                    <Button label="Delete" icon="pi pi-trash" severity="secondary" @click="confirmDeleteSelected" :disabled="!selectedAgents || !selectedAgents.length" />
                </template>

                <template #end>
                    <Button label="Export" icon="pi pi-upload" severity="secondary" @click="exportCSV($event)" />
                </template>
            </Toolbar>

            <DataTable
                ref="dt"
                v-model:selection="selectedAgents"
                :value="agents"
                dataKey="id"
                :paginator="true"
                :rows="10"
                :filters="filters"
                :loading="loading"
                paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
                :rowsPerPageOptions="[5, 10, 25]"
                currentPageReportTemplate="Showing {first} to {last} of {totalRecords} agents"
            >
                <template #header>
                    <div class="flex flex-wrap gap-2 items-center justify-between">
                        <h4 class="m-0">Manage Agents</h4>
                        <IconField>
                            <InputIcon>
                                <i class="pi pi-search" />
                            </InputIcon>
                            <InputText v-model="filters['global'].value" placeholder="Search..." />
                        </IconField>
                    </div>
                </template>

                <template #empty>
                    <div class="text-center p-4">No agents found.</div>
                </template>

                <template #loading>
                    <div class="text-center p-4">Loading agents...</div>
                </template>

                <Column selectionMode="multiple" style="width: 3rem" :exportable="false"></Column>
                <Column field="id" header="ID" sortable style="width: 5rem"></Column>
                <Column field="name" header="Name" sortable style="min-width: 12rem"></Column>
                <Column field="os" header="OS" sortable style="min-width: 8rem">
                    <template #body="slotProps">
                        <div class="flex items-center gap-2">
                            <i class="pi" :class="{
                                'pi-microsoft': slotProps.data.os === 'windows',
                                'pi-apple': slotProps.data.os === 'macos',
                                'pi-desktop': slotProps.data.os === 'linux'
                            }"></i>
                            <span>{{ slotProps.data.os }}</span>
                        </div>
                    </template>
                </Column>
                <Column field="status" header="Status" sortable style="min-width: 8rem">
                    <template #body="slotProps">
                        <Tag :value="slotProps.data.status" :severity="getStatusSeverity(slotProps.data.status)" />
                    </template>
                </Column>
                <Column field="lastSeen" header="Last Seen" sortable style="min-width: 10rem">
                    <template #body="slotProps">
                        {{ formatDate(slotProps.data.lastSeen) }}
                    </template>
                </Column>
                <Column field="description" header="Description" sortable style="min-width: 16rem"></Column>
                <Column :exportable="false" style="min-width: 16rem">
                    <template #body="slotProps">
                        <Button icon="pi pi-download" outlined rounded class="mr-2" @click="downloadAgent(slotProps.data)" tooltip="Download Agent" tooltipPosition="top" />
                        <Button icon="pi pi-pencil" outlined rounded class="mr-2" @click="editAgent(slotProps.data)" />
                        <Button icon="pi pi-trash" outlined rounded severity="danger" @click="confirmDeleteAgent(slotProps.data)" />
                    </template>
                </Column>
            </DataTable>
        </div>

        <Dialog v-model:visible="agentDialog" :style="{ width: '450px' }" header="Agent Details" :modal="true">
            <div class="flex flex-col gap-6">
                <div>
                    <label for="name" class="block font-bold mb-3">Name</label>
                    <InputText id="name" v-model.trim="agent.name" required="true" autofocus :class="{ 'p-invalid': submitted && !agent.name }" fluid />
                    <small v-if="submitted && !agent.name" class="text-red-500">Name is required.</small>
                </div>
                
                <div>
                    <label for="os" class="block font-bold mb-3">Operating System</label>
                    <Select id="os" v-model="agent.os" :options="operatingSystems" optionLabel="label" placeholder="Select an OS" :class="{ 'p-invalid': submitted && !agent.os }" fluid></Select>
                    <small v-if="submitted && !agent.os" class="text-red-500">OS is required.</small>
                </div>
                
                <div>
                    <label for="status" class="block font-bold mb-3">Status</label>
                    <Select id="status" v-model="agent.status" :options="statuses" optionLabel="label" placeholder="Select a Status" fluid></Select>
                </div>
                
                <div>
                    <label for="description" class="block font-bold mb-3">Description</label>
                    <Textarea id="description" v-model="agent.description" rows="3" cols="20" fluid />
                </div>
            </div>

            <template #footer>
                <Button label="Cancel" icon="pi pi-times" text @click="hideDialog" />
                <Button label="Save" icon="pi pi-check" @click="saveAgent" />
            </template>
        </Dialog>

        <Dialog v-model:visible="deleteAgentDialog" :style="{ width: '450px' }" header="Confirm" :modal="true">
            <div class="flex items-center gap-4">
                <i class="pi pi-exclamation-triangle !text-3xl" />
                <span v-if="agent">
                    Are you sure you want to delete <b>{{ agent.name }}</b>?
                </span>
            </div>
            <template #footer>
                <Button label="No" icon="pi pi-times" text @click="deleteAgentDialog = false" />
                <Button label="Yes" icon="pi pi-check" @click="deleteAgent" />
            </template>
        </Dialog>

        <Dialog v-model:visible="deleteAgentsDialog" :style="{ width: '450px' }" header="Confirm" :modal="true">
            <div class="flex items-center gap-4">
                <i class="pi pi-exclamation-triangle !text-3xl" />
                <span>Are you sure you want to delete the selected agents?</span>
            </div>
            <template #footer>
                <Button label="No" icon="pi pi-times" text @click="deleteAgentsDialog = false" />
                <Button label="Yes" icon="pi pi-check" @click="deleteSelectedAgents" />
            </template>
        </Dialog>
    </div>
</template>