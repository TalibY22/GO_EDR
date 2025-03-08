<template>
  <div>
    <b-container fluid class="mt-4">
      <b-row>
        <b-col>
          <b-card class="shadow">
            <b-card-header class="d-flex justify-content-between align-items-center">
              <h3 class="mb-0">Bash Command History</h3>
              <div>
                <b-button variant="outline-primary" size="sm" @click="refreshHistory">
                  <b-icon icon="arrow-clockwise"></b-icon> Refresh
                </b-button>
              </div>
            </b-card-header>
            <b-card-body>
              <b-row>
                <b-col md="4">
                  <b-form-group label="Agent ID">
                    <b-form-input v-model="filters.agentId" placeholder="Filter by Agent ID"></b-form-input>
                  </b-form-group>
                </b-col>
                <b-col md="4">
                  <b-form-group label="Search Commands">
                    <b-form-input v-model="filters.command" placeholder="Search in commands"></b-form-input>
                  </b-form-group>
                </b-col>
                <b-col md="4">
                  <b-form-group label="Date Range">
                    <b-form-datepicker v-model="filters.date" placeholder="Select Date"></b-form-datepicker>
                  </b-form-group>
                </b-col>
              </b-row>
              <b-button variant="primary" @click="applyFilters">Apply Filters</b-button>
              <b-button variant="secondary" class="ml-2" @click="resetFilters">Reset</b-button>
            </b-card-body>
          </b-card>
        </b-col>
      </b-row>

      <!-- Bash History Table -->
      <b-row class="mt-4">
        <b-col>
          <b-card class="shadow">
            <div class="table-responsive">
              <b-table 
                striped 
                hover 
                bordered 
                :items="filteredHistory" 
                :fields="fields" 
                class="align-items-center table-flush"
                :sort-by.sync="sortBy"
                :sort-desc.sync="sortDesc"
                :busy="isLoading"
              >
                <template #table-busy>
                  <div class="text-center my-3">
                    <b-spinner variant="primary" label="Loading..."></b-spinner>
                    <div class="mt-2">Loading bash history...</div>
                  </div>
                </template>
                <template #cell(details)="row">
                  <code>{{ row.item.details }}</code>
                </template>
                <template #cell(actions)="row">
                  <b-button 
                    size="sm" 
                    variant="primary" 
                    @click="reissueCommand(row.item)"
                    title="Re-issue this command"
                  >
                    <b-icon icon="terminal"></b-icon>
                  </b-button>
                </template>
              </b-table>
            </div>
            <b-pagination 
              v-model="currentPage" 
              :total-rows="filteredHistory.length" 
              :per-page="perPage" 
              class="mt-3 justify-content-center"
            ></b-pagination>
          </b-card>
        </b-col>
      </b-row>
    </b-container>
  </div>
</template>

<script>
export default {
  data() {
    return {
      bashHistory: [],
      isLoading: false,
      filters: {
        agentId: '',
        command: '',
        date: null
      },
      fields: [
        { key: 'id', label: 'ID', sortable: true },
        { key: 'agent_id', label: 'Agent ID', sortable: true },
        { key: 'timestamp', label: 'Timestamp', sortable: true },
        { key: 'details', label: 'Command' },
        { key: 'actions', label: 'Actions' }
      ],
      currentPage: 1,
      perPage: 15,
      sortBy: 'timestamp',
      sortDesc: true,
      showDetailsModal: false,
      selectedCommand: null
    };
  },
  computed: {
    filteredHistory() {
      let filtered = [...this.bashHistory];
      
      // Apply agent ID filter
      if (this.filters.agentId) {
        filtered = filtered.filter(item => 
          item.agent_id.toLowerCase().includes(this.filters.agentId.toLowerCase())
        );
      }
      
      // Apply command filter
      if (this.filters.command) {
        filtered = filtered.filter(item => 
          item.details.toLowerCase().includes(this.filters.command.toLowerCase())
        );
      }
      
      // Apply date filter
      if (this.filters.date) {
        const filterDate = new Date(this.filters.date);
        filtered = filtered.filter(item => {
          const itemDate = new Date(item.timestamp);
          return itemDate.toDateString() === filterDate.toDateString();
        });
      }
      
      return filtered;
    }
  },
  methods: {
    async fetchBashHistory() {
      this.isLoading = true;
      try {
        const response = await fetch('http://localhost:8080/bash-history');
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        this.bashHistory = data.data || [];
      } catch (error) {
        console.error('Error fetching bash history:', error);
        this.$bvToast.toast(`Failed to fetch bash history: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      } finally {
        this.isLoading = false;
      }
    },
    refreshHistory() {
      this.fetchBashHistory();
    },
    applyFilters() {
      this.currentPage = 1; // Reset pagination when applying filters
    },
    resetFilters() {
      this.filters = { agentId: '', command: '', date: null };
      this.currentPage = 1;
    },
    async reissueCommand(historyItem) {
      try {
        // Create a new command to be executed
        const response = await fetch('http://localhost:8080/command', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            agent_id: historyItem.agent_id,
            command: historyItem.details
          })
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Show success message
        this.$bvToast.toast(`Command "${historyItem.details}" re-issued to agent ${historyItem.agent_id}`, {
          title: 'Command Sent',
          variant: 'success',
          solid: true
        });
        
        // Redirect to commands page to see the result
        this.$router.push('/commands');
      } catch (error) {
        console.error('Error re-issuing command:', error);
        this.$bvToast.toast(`Failed to re-issue command: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      }
    }
  },
  mounted() {
    this.fetchBashHistory();
    
    // Refresh history every minute
    this.refreshInterval = setInterval(() => {
      this.fetchBashHistory();
    }, 60000);
  },
  beforeDestroy() {
    // Clear the interval when component is destroyed
    clearInterval(this.refreshInterval);
  }
};
</script>

<style>
code {
  background-color: #f8f9fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}
</style> 