<template>
  <div>
    <!-- Stats Card -->
   
    <!-- Filters -->
    <b-container fluid class="mt-4">
      <b-row>
        <b-col>
          <b-card class="shadow">
            <b-card-header>
              <h3 class="mb-0">Log Filters</h3>
            </b-card-header>
            <b-card-body>
              <b-row>
                <b-col md="3">
                  <b-form-group label="Agent ID">
                    <b-form-input v-model="filters.agentId" placeholder="Search by Agent ID"></b-form-input>
                  </b-form-group>
                </b-col>
                <b-col md="3">
                  <b-form-group label="Event Type">
                    <b-form-select v-model="filters.eventType" :options="eventTypes" placeholder="Select Event Type"></b-form-select>
                  </b-form-group>
                </b-col>
                <b-col md="3">
                  <b-form-group label="Date Range">
                    <b-form-datepicker v-model="filters.date" placeholder="Select Date"></b-form-datepicker>
                  </b-form-group>
                </b-col>
                <b-col md="3">
                  <b-form-group label="Search Details">
                    <b-form-input v-model="filters.details" placeholder="Search in details"></b-form-input>
                  </b-form-group>
                </b-col>
              </b-row>
              <b-button variant="primary" @click="applyFilters">Apply Filters</b-button>
              <b-button variant="secondary" class="ml-2" @click="resetFilters">Reset</b-button>
            </b-card-body>
          </b-card>
        </b-col>
      </b-row>

      <!-- Logs Table -->
      <b-row class="mt-4">
        <b-col>
          <b-card class="shadow">
            <b-card-header class="d-flex justify-content-between align-items-center">
              <h3 class="mb-0">Logs</h3>
              <div>
                <b-button variant="outline-primary" size="sm" @click="refreshLogs">
                  <b-icon icon="arrow-clockwise"></b-icon> Refresh
                </b-button>
                <b-button variant="outline-danger" size="sm" class="ml-2" @click="clearAllLogs">
                  <b-icon icon="trash"></b-icon> Clear All
                </b-button>
              </div>
            </b-card-header>
            <div class="table-responsive">
              <b-table 
                striped 
                hover 
                bordered 
                :items="filteredLogs" 
                :fields="fields" 
                class="align-items-center table-flush"
                :sort-by.sync="sortBy"
                :sort-desc.sync="sortDesc"
                :busy="isLoading"
              >
                <template #table-busy>
                  <div class="text-center my-3">
                    <b-spinner variant="primary" label="Loading..."></b-spinner>
                    <div class="mt-2">Loading logs...</div>
                  </div>
                </template>
                <template #cell(event)="row">
                  <b-badge :variant="getEventVariant(row.item.event)">{{ row.item.event }}</b-badge>
                </template>
                <template #cell(details)="row">
                  <div class="log-details">{{ row.item.details }}</div>
                  <b-button 
                    v-if="row.item.details.length > 100" 
                    size="sm" 
                    variant="link" 
                    @click="viewFullDetails(row.item)"
                  >
                    View Full Details
                  </b-button>
                </template>
                <template #cell(actions)="row">
                  <b-button size="sm" variant="danger" @click="deleteLog(row.item.id)">
                    <b-icon icon="trash"></b-icon>
                  </b-button>
                </template>
              </b-table>
            </div>
            <b-pagination 
              v-model="currentPage" 
              :total-rows="filteredLogs.length" 
              :per-page="perPage" 
              class="mt-3 justify-content-center"
            ></b-pagination>
          </b-card>
        </b-col>
      </b-row>
    </b-container>

    <!-- Details Modal -->
    <b-modal v-model="showDetailsModal" title="Log Details" size="lg" ok-only>
      <pre class="log-details-modal">{{ selectedLogDetails }}</pre>
    </b-modal>
  </div>
</template>

<script>
export default {
  data() {
    return {
      logs: [],
      isLoading: false,
      filters: {
        agentId: '',
        eventType: '',
        date: null,
        details: ''
      },
      eventTypes: [
        { value: '', text: 'All Events' },
        { value: 'PROCESS', text: 'Process' },
        { value: 'FILE', text: 'File' },
        { value: 'NETWORK', text: 'Network' },
        { value: 'REGISTRY', text: 'Registry' },
        { value: 'DNS', text: 'DNS' },
        { value: 'BASH', text: 'Bash History' },
        { value: 'ERROR', text: 'Error' }
      ],
      fields: [
        { key: 'id', label: 'ID', sortable: true },
        { key: 'agent_id', label: 'Agent ID', sortable: true },
        { key: 'timestamp', label: 'Timestamp', sortable: true },
        { key: 'event', label: 'Event Type', sortable: true },
        { key: 'details', label: 'Details' },
        { key: 'actions', label: 'Actions' }
      ],
      currentPage: 1,
      perPage: 15,
      sortBy: 'timestamp',
      sortDesc: true,
      showDetailsModal: false,
      selectedLogDetails: ''
    };
  },
  computed: {
    filteredLogs() {
      return this.logs.filter(log => 
        (!this.filters.agentId || log.agent_id.toString().includes(this.filters.agentId)) &&
        (!this.filters.eventType || log.event === this.filters.eventType) &&
        (!this.filters.date || (log.timestamp && log.timestamp.includes(this.filters.date))) &&
        (!this.filters.details || (log.details && log.details.toLowerCase().includes(this.filters.details.toLowerCase())))
      );
    }
  },
  methods: {
    async fetchLogs() {
      this.isLoading = true;
      try {
        const response = await fetch('http://localhost:8080/logs');
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        this.logs = data.data || [];
      } catch (error) {
        console.error('Error fetching logs:', error);
        this.$bvToast.toast(`Failed to fetch logs: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      } finally {
        this.isLoading = false;
      }
    },
    refreshLogs() {
      this.fetchLogs();
    },
    applyFilters() {
      this.currentPage = 1; // Reset pagination when applying filters
    },
    resetFilters() {
      this.filters = { agentId: '', eventType: '', date: null, details: '' };
      this.currentPage = 1;
    },
    async deleteLog(id) {
      try {
        const response = await fetch(`http://localhost:8080/logs/${id}`, {
          method: 'DELETE'
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Remove the log from the list
        this.logs = this.logs.filter(log => log.id !== id);
        
        this.$bvToast.toast('Log deleted successfully', {
          title: 'Success',
          variant: 'success',
          solid: true
        });
      } catch (error) {
        console.error('Error deleting log:', error);
        this.$bvToast.toast(`Failed to delete log: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      }
    },
    async clearAllLogs() {
      if (!confirm('Are you sure you want to delete all logs? This action cannot be undone.')) {
        return;
      }
      
      try {
        const response = await fetch('http://localhost:8080/logs/clear', {
          method: 'DELETE'
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Clear the logs array
        this.logs = [];
        
        this.$bvToast.toast('All logs cleared successfully', {
          title: 'Success',
          variant: 'success',
          solid: true
        });
      } catch (error) {
        console.error('Error clearing logs:', error);
        this.$bvToast.toast(`Failed to clear logs: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      }
    },
    viewFullDetails(log) {
      this.selectedLogDetails = log.details;
      this.showDetailsModal = true;
    },
    getEventVariant(event) {
      switch (event) {
        case 'PROCESS':
          return 'primary';
        case 'FILE':
          return 'success';
        case 'NETWORK':
          return 'info';
        case 'REGISTRY':
          return 'warning';
        case 'DNS':
          return 'secondary';
        case 'BASH':
          return 'dark';
        case 'ERROR':
          return 'danger';
        default:
          return 'light';
      }
    }
  },
  mounted() {
    this.fetchLogs();
    
    // Refresh logs every minute
    this.refreshInterval = setInterval(() => {
      this.fetchLogs();
    }, 60000);
  },
  beforeDestroy() {
    // Clear the interval when component is destroyed
    clearInterval(this.refreshInterval);
  }
};
</script>

<style>
.log-details {
  max-height: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.log-details-modal {
  white-space: pre-wrap;
  max-height: 500px;
  overflow-y: auto;
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 5px;
}
</style>
