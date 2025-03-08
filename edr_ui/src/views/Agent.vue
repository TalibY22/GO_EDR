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
            <b-table striped hover bordered :items="agents" :fields="fields" class="table align-items-center table-flush">
              <template #cell(status)="row">
                <b-badge :variant="getStatusVariant(row.item.status)">{{ row.item.status }}</b-badge>
              </template>
              <template #cell(actions)="row">
                <b-button size="sm" variant="info" class="mr-2" @click="downloadAgent(row.item)">
                  <b-icon icon="download"></b-icon> Download
                </b-button>
                <b-button size="sm" variant="danger" @click="deleteAgent(row.item.id)">
                  <b-icon icon="trash"></b-icon> Delete
                </b-button>
              </template>
            </b-table>
          </b-card>
        </b-col>
      </b-row>
    </b-container>

    <!-- Modal Form -->
    <b-modal v-model="isModalOpen" title="Generate New Agent" hide-footer>
      <b-form @submit.prevent="generateAgent">
        <b-form-group label="Agent Name">
          <b-form-input v-model="newAgent.name" required placeholder="Enter agent name"></b-form-input>
        </b-form-group>
        <b-form-group label="Operating System">
          <b-form-select v-model="newAgent.os" :options="osOptions" required></b-form-select>
        </b-form-group>
        <b-form-group label="Description">
          <b-form-textarea v-model="newAgent.description" placeholder="Enter description (optional)"></b-form-textarea>
        </b-form-group>
        <div class="d-flex justify-content-end">
          <b-button variant="secondary" @click="closeModal">Cancel</b-button>
          <b-button type="submit" variant="primary" class="ml-2" :disabled="isGenerating">
            <b-spinner small v-if="isGenerating"></b-spinner>
            Generate
          </b-button>
        </div>
      </b-form>
    </b-modal>
  </div>
</template>

<script>
export default {
  data() {
    return {
      agents: [],
      newAgent: {
        name: '',
        os: 'linux',
        description: ''
      },
      isModalOpen: false,
      isGenerating: false,
      fields: [
        { key: 'id', label: 'ID', sortable: true },
        { key: 'name', label: 'Name', sortable: true },
        { key: 'os', label: 'OS', sortable: true },
        { key: 'status', label: 'Status', sortable: true },
        { key: 'last_seen', label: 'Last Seen', sortable: true },
        { key: 'description', label: 'Description' },
        { key: 'actions', label: 'Actions' }
      ],
      osOptions: [
        { value: 'linux', text: 'Linux' },
        { value: 'windows', text: 'Windows' },
        { value: 'macos', text: 'macOS' }
      ]
    };
  },
  methods: {
    async fetchAgents() {
      try {
        const response = await fetch('http://localhost:8080/agents');
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        this.agents = data.data || [];
      } catch (error) {
        console.error('Error fetching agents:', error);
      }
    },
    async generateAgent() {
      this.isGenerating = true;
      
      try {
        const response = await fetch('http://localhost:8080/agents', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            name: this.newAgent.name,
            os: this.newAgent.os,
            description: this.newAgent.description,
            status: 'Pending'
          })
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        
        // Add the new agent to the list
        this.agents.push(data.data);
        
        // Reset form and close modal
        this.newAgent = {
          name: '',
          os: 'linux',
          description: ''
        };
        this.closeModal();
        
        // Show success message
        this.$bvToast.toast('Agent generated successfully. You can now download it.', {
          title: 'Success',
          variant: 'success',
          solid: true
        });
      } catch (error) {
        console.error('Error generating agent:', error);
        
        // Show error message
        this.$bvToast.toast(`Failed to generate agent: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      } finally {
        this.isGenerating = false;
      }
    },
    async deleteAgent(id) {
      if (!confirm('Are you sure you want to delete this agent?')) {
        return;
      }
      
      try {
        const response = await fetch(`http://localhost:8080/agents/${id}`, {
          method: 'DELETE'
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Remove the agent from the list
        this.agents = this.agents.filter(agent => agent.id !== id);
        
        // Show success message
        this.$bvToast.toast('Agent deleted successfully.', {
          title: 'Success',
          variant: 'success',
          solid: true
        });
      } catch (error) {
        console.error('Error deleting agent:', error);
        
        // Show error message
        this.$bvToast.toast(`Failed to delete agent: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      }
    },
    async downloadAgent(agent) {
      try {
        const response = await fetch(`http://localhost:8080/agents/${agent.id}/download`);
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Create a blob from the response
        const blob = await response.blob();
        
        // Create a link to download the file
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = `edr-agent-${agent.name.toLowerCase().replace(/\s+/g, '-')}`;
        
        // Add the link to the document and click it
        document.body.appendChild(a);
        a.click();
        
        // Clean up
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      } catch (error) {
        console.error('Error downloading agent:', error);
        
        // Show error message
        this.$bvToast.toast(`Failed to download agent: ${error.message}`, {
          title: 'Error',
          variant: 'danger',
          solid: true
        });
      }
    },
    getStatusVariant(status) {
      switch (status) {
        case 'Active':
          return 'success';
        case 'Pending':
          return 'warning';
        case 'Offline':
          return 'danger';
        default:
          return 'secondary';
      }
    },
    openModal() {
      this.isModalOpen = true;
    },
    closeModal() {
      this.isModalOpen = false;
    }
  },
  mounted() {
    this.fetchAgents();
    
    // Refresh agents list every 30 seconds
    setInterval(() => {
      this.fetchAgents();
    }, 30000);
  }
};
</script>
