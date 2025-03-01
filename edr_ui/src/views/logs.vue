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
                <b-col md="4">
                  <b-form-group label="Agent Name">
                    <b-form-input v-model="filters.agentName" placeholder="Search by Agent Name"></b-form-input>
                  </b-form-group>
                </b-col>
                <b-col md="4">
                  <b-form-group label="Log Type">
                    <b-form-select v-model="filters.logType" :options="logTypes" placeholder="Select Log Type"></b-form-select>
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

      <!-- Logs Table -->
      <b-row class="mt-4">
        <b-col>
          <b-card class="shadow">
            <b-card-header>
              <h3 class="mb-0">Logs</h3>
            </b-card-header>
            <b-table striped hover bordered :items="filteredLogs" :fields="fields" class="align-items-center table-flush">
              <template #cell(actions)="row">
                <b-button size="sm" variant="danger" @click="deleteLog(row.item.id)">Delete</b-button>
              </template>
            </b-table>
            <b-pagination v-model="currentPage" :total-rows="filteredLogs.length" :per-page="perPage" class="mt-3"></b-pagination>
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
      logs: [],
      filters: {
        agentName: '',
        logType: '',
        date: null
      },
      logTypes: ["Error", "Warning", "Info", "Debug"],
      fields: [
        { key: 'id', label: 'ID', sortable: true },
        { key: 'agent', label: 'Agent Name', sortable: true },
        { key: 'logType', label: 'Log Type', sortable: true },
        { key: 'message', label: 'Message' },
        { key: 'timestamp', label: 'Timestamp', sortable: true },
        { key: 'actions', label: 'Actions' }
      ],
      currentPage: 1,
      perPage: 10
    };
  },
  computed: {
    filteredLogs() {
      return this.logs.filter(log => 
        (!this.filters.agentName || log.agent.toLowerCase().includes(this.filters.agentName.toLowerCase())) &&
        (!this.filters.logType || log.logType === this.filters.logType) &&
        (!this.filters.date || log.timestamp.startsWith(this.filters.date))
      );
    }
  },
  methods: {
    async fetchLogs() {
      // Replace with actual API call
      this.logs = [
        { id: 1, agent: "Agent-001", logType: "Error", message: "System crash", timestamp: "2025-02-10 12:34:56" },
        { id: 2, agent: "Agent-002", logType: "Warning", message: "High memory usage", timestamp: "2025-02-10 12:45:00" },
        { id: 3, agent: "Agent-003", logType: "Info", message: "Heartbeat received", timestamp: "2025-02-10 13:00:10" }
      ];
    },
    applyFilters() {
      this.currentPage = 1; // Reset pagination when applying filters
    },
    resetFilters() {
      this.filters = { agentName: '', logType: '', date: null };
    },
    async deleteLog(id) {
      this.logs = this.logs.filter(log => log.id !== id);
    }
  },
  mounted() {
    this.fetchLogs();
  }
};
</script>
