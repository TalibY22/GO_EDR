<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';

const logs = ref([]);
const isConnected = ref(false);
const connectionStatus = ref('Disconnected');
const autoScroll = ref(true);
const maxLogs = ref(100); // Maximum number of logs to keep in memory
const pollingInterval = ref(2000); // Polling interval in milliseconds
const eventFilter = ref('');
const agentFilter = ref('');

let pollingTimer = null;

// Function to fetch the latest logs
const fetchLatestLogs = async () => {
    try {
        // Get the latest log ID to fetch only newer logs
        const lastLogId = logs.value.length > 0 ? logs.value[0].id : 0;
        
        // Build query parameters
        let queryParams = `?since_id=${lastLogId}`;
        if (eventFilter.value) queryParams += `&event=${eventFilter.value}`;
        if (agentFilter.value) queryParams += `&agent_id=${agentFilter.value}`;
        
        const response = await fetch(`http://localhost:8080/logs/latest${queryParams}`);
        const data = await response.json();
        
        if (data.data && data.data.length > 0) {
            // Add new logs to the beginning of the array
            logs.value = [...data.data, ...logs.value];
            
            // Limit the number of logs to prevent memory issues
            if (logs.value.length > maxLogs.value) {
                logs.value = logs.value.slice(0, maxLogs.value);
            }
            
            // Auto-scroll to the latest log
            if (autoScroll.value) {
                scrollToLatest();
            }
        }
        
        if (!isConnected.value) {
            isConnected.value = true;
            connectionStatus.value = 'Connected';
        }
    } catch (error) {
        console.error('Error fetching logs:', error);
        isConnected.value = false;
        connectionStatus.value = 'Connection Error';
    }
};

// Start polling for new logs
const startPolling = () => {
    isConnected.value = true;
    connectionStatus.value = 'Connecting...';
    
    // Clear any existing timer
    if (pollingTimer) {
        clearInterval(pollingTimer);
    }
    
    // Initial fetch
    fetchLatestLogs();
    
    // Set up polling interval
    pollingTimer = setInterval(fetchLatestLogs, pollingInterval.value);
};

// Stop polling
const stopPolling = () => {
    if (pollingTimer) {
        clearInterval(pollingTimer);
        pollingTimer = null;
    }
    isConnected.value = false;
    connectionStatus.value = 'Disconnected';
};

// Toggle connection
const toggleConnection = () => {
    if (isConnected.value) {
        stopPolling();
    } else {
        startPolling();
    }
};

// Clear logs
const clearLogs = () => {
    logs.value = [];
};

// Scroll to the latest log
const scrollToLatest = () => {
    setTimeout(() => {
        const container = document.querySelector('.log-container');
        if (container) {
            container.scrollTop = 0;
        }
    }, 100);
};

// Apply filters
const applyFilters = () => {
    logs.value = []; // Clear existing logs
    fetchLatestLogs(); // Fetch with new filters
};

// Get color based on event type
const getEventColor = (event) => {
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

// Format timestamp
const formatTimestamp = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString();
};

// Start polling when component is mounted
onMounted(() => {
    startPolling();
});

// Clean up when component is unmounted
onBeforeUnmount(() => {
    stopPolling();
});
</script>

<template>
    <div class="card">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">Live Log Monitor</h1>
            <div class="flex items-center">
                <span class="mr-2">Status:</span>
                <span 
                    class="px-2 py-1 rounded-full text-sm font-semibold"
                    :class="isConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                >
                    {{ connectionStatus }}
                </span>
            </div>
        </div>
        
        <div class="grid grid-cols-12 gap-4 mb-4">
            <div class="col-span-3">
                <div class="flex flex-col">
                    <label class="mb-1 text-sm">Event Type Filter</label>
                    <InputText v-model="eventFilter" placeholder="Filter by event type" />
                </div>
            </div>
            <div class="col-span-3">
                <div class="flex flex-col">
                    <label class="mb-1 text-sm">Agent ID Filter</label>
                    <InputText v-model="agentFilter" placeholder="Filter by agent ID" />
                </div>
            </div>
            <div class="col-span-2">
                <div class="flex flex-col">
                    <label class="mb-1 text-sm">Max Logs</label>
                    <InputNumber v-model="maxLogs" :min="10" :max="1000" />
                </div>
            </div>
            <div class="col-span-2">
                <div class="flex flex-col">
                    <label class="mb-1 text-sm">Poll Interval (ms)</label>
                    <InputNumber v-model="pollingInterval" :min="500" :max="10000" :step="500" />
                </div>
            </div>
            <div class="col-span-2 flex items-end">
                <Button label="Apply Filters" @click="applyFilters" class="mr-2" />
            </div>
        </div>
        
        <div class="flex justify-between mb-4">
            <div>
                <Button 
                    :label="isConnected ? 'Disconnect' : 'Connect'" 
                    :class="isConnected ? 'p-button-danger' : 'p-button-success'" 
                    @click="toggleConnection" 
                    class="mr-2"
                />
                <Button label="Clear Logs" @click="clearLogs" class="p-button-secondary" />
            </div>
            <div class="flex items-center">
                <Checkbox v-model="autoScroll" :binary="true" id="auto-scroll" />
                <label for="auto-scroll" class="ml-2">Auto-scroll to new logs</label>
            </div>
        </div>
        
        <div class="log-container border rounded-lg h-[600px] overflow-y-auto">
            <div v-if="logs.length === 0" class="flex justify-center items-center h-full text-gray-500">
                No logs available. Waiting for new logs...
            </div>
            <div v-else class="divide-y">
                <div 
                    v-for="log in logs" 
                    :key="log.id" 
                    class="p-3 hover:bg-gray-50 transition-colors"
                >
                    <div class="flex justify-between">
                        <div class="flex items-center">
                            <span 
                                class="px-2 py-1 rounded-full text-xs font-semibold mr-2"
                                :class="getEventColor(log.event)"
                            >
                                {{ log.event }}
                            </span>
                            <span class="text-sm text-gray-600">Agent ID: {{ log.agent_id }}</span>
                        </div>
                        <div class="text-sm text-gray-500">
                            {{ formatTimestamp(log.created_at) }}
                        </div>
                    </div>
                    <div class="mt-2 text-sm whitespace-pre-wrap">{{ log.details }}</div>
                </div>
            </div>
        </div>
        
        <div class="mt-4 text-sm text-gray-500">
            Showing {{ logs.length }} logs (max: {{ maxLogs }})
        </div>
    </div>
</template>

<style scoped>
.log-container {
    /* Reverse the scroll direction to show newest logs at the top */
    display: flex;
    flex-direction: column-reverse;
}
</style> 