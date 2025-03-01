<template>
  <div>
    <!-- Command Terminal -->
    <b-container fluid class="mt-4">
      <b-row>
        <b-col>
          <b-card class="shadow">
            <b-card-header>
              <h3 class="mb-0">Command Terminal</h3>
            </b-card-header>
            <b-card-body class="terminal p-3">
              <div class="terminal-window">
                <div v-for="(cmd, index) in commands" :key="index" class="terminal-line">
                  <span class="text-info"># {{ cmd.command }}</span>
                  <pre class="text-blue">{{ cmd.response }}</pre>
                </div>
              </div>
              <b-form-input v-model="newCommand" placeholder="Enter command..." class="mt-3" @keyup.enter="executeCommand"></b-form-input>
              <b-button variant="primary" class="mt-2" @click="executeCommand">Run</b-button>
            </b-card-body>
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
      commands: [],
      newCommand: ""
    };
  },
  methods: {
    async executeCommand() {
      if (!this.newCommand.trim()) return;

      const commandText = this.newCommand;
      this.newCommand = "";

      // Simulate API call to execute command
      const response = await this.sendCommandToAgent(commandText);

      // Store command and response
      this.commands.push({ command: commandText, response });
    },
    async sendCommandToAgent(command) {
      // Replace with actual API request to execute command
      return `Executed: ${command}\nResponse: OK`;
    }
  }
};
</script>

<style>
.terminal {
  background-color: black;
  color: #0f0;
  font-family: monospace;
  padding: 20px;
  border-radius: 8px;
  min-height: 300px;
  overflow-y: auto;
}
.terminal-window {
  max-height: 300px;
  overflow-y: auto;
}
.terminal-line {
  margin-bottom: 8px;
}
</style>
