<template>
  <div>
    <!-- EDR Agent Table -->
    <b-container fluid class="mt-4">
      <b-row>
        <b-col>
          <b-card no-body class="shadow">
            <b-card-header class="border-0 d-flex justify-content-between align-items-center">
              <h3 class="mb-0">Agent List</h3>
              <b-button variant="primary" @click="openModal">+ Add Agent</b-button>
            </b-card-header>
            <b-table striped hover bordered :items="edrs" :fields="fields" class="table align-items-center table-flush">
              <template #cell(actions)="row">
                <b-button size="sm" variant="danger" @click="deleteEDR(row.item.id)">Delete</b-button>
              </template>
            </b-table>
          </b-card>
        </b-col>
      </b-row>
    </b-container>

    <!-- Modal Form -->
    <b-modal v-model="isModalOpen" title="Generate New EDR" hide-footer>
      <b-form @submit.prevent="generateEDR">
        <b-form-group label="EDR Name">
          <b-form-input v-model="newEdrName" required placeholder="Enter EDR name"></b-form-input>
        </b-form-group>
        <div class="d-flex justify-content-end">
          <b-button variant="secondary" @click="closeModal">Cancel</b-button>
          <b-button type="submit" variant="primary" class="ml-2">Generate</b-button>
        </div>
      </b-form>
    </b-modal>
  </div>
</template>

<script>
export default {
  data() {
    return {
      edrs: [],
      newEdrName: '',
      isModalOpen: false,
      fields: [
        { key: 'id', label: 'ID', sortable: true },
        { key: 'name', label: 'Name', sortable: true },
        { key: 'status', label: 'Status', sortable: true },
        { key: 'actions', label: 'Actions', sortable: false }
      ]
    };
  },
  methods: {
    async fetchEDRs() {
      // Replace with API call
      this.edrs = [
        { id: 1, name: "EDR-001", status: "Active" },
        { id: 2, name: "EDR-002", status: "Pending" }
      ];
    },
    async generateEDR() {
      if (!this.newEdrName) return;
      this.edrs.push({ id: this.edrs.length + 1, name: this.newEdrName, status: "New" });
      this.newEdrName = '';
      this.closeModal();
    },
    async deleteEDR(id) {
      this.edrs = this.edrs.filter(edr => edr.id !== id);
    },
    openModal() {
      this.isModalOpen = true;
    },
    closeModal() {
      this.isModalOpen = false;
    }
  },
  mounted() {
    this.fetchEDRs();
  }
};
</script>
