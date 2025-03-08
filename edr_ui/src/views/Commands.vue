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
                  <div v-if="cmd.status === 'pending'" class="text-warning">
                    <b-spinner small variant="warning" class="mr-2"></b-spinner>
                    Command issued, waiting for response...
                  </div>
                  <div v-else-if="cmd.status === 'interactive'" class="text-warning">
                    <pre v-if="cmd.partialResponse" class="text-blue">{{ cmd.partialResponse }}</pre>
                    <div class="interactive-input">
                      <b-form-input 
                        v-model="cmd.interactiveInput" 
                        :placeholder="cmd.inputPrompt || 'Enter input...'" 
                        :type="cmd.inputType || 'text'"
                        @keyup.enter="sendInteractiveInput(cmd)"
                        autofocus
                      ></b-form-input>
                      <b-button 
                        variant="primary" 
                        size="sm" 
                        class="ml-2" 
                        @click="sendInteractiveInput(cmd)"
                      >
                        Send
                      </b-button>
                    </div>
                  </div>
                  <div v-else>
                    <pre class="text-blue">{{ cmd.response }}</pre>
                    <!-- Show link to files view if this was a send command -->
                    <div v-if="cmd.command === 'send' && cmd.status === 'complete'" class="mt-2">
                      <b-button variant="success" size="sm" @click="goToFiles">
                        View Uploaded Files
                      </b-button>
                    </div>
                  </div>
                </div>
              </div>
              <b-form-input 
                v-model="newCommand" 
                placeholder="Enter command..." 
                class="mt-3" 
                @keyup.enter="executeCommand"
                :disabled="loading || hasInteractiveCommand"
              ></b-form-input>
              <b-button 
                variant="primary" 
                class="mt-2" 
                @click="executeCommand"
                :disabled="loading || hasInteractiveCommand"
              >
                <b-spinner small v-if="loading" class="mr-1"></b-spinner>
                Run
              </b-button>
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
      newCommand: "",
      activeCommand: null,
      loading: false,
      pollingInterval: null
    };
  },
  computed: {
    hasInteractiveCommand() {
      return this.commands.some(cmd => cmd.status === 'interactive');
    }
  },
  created() {
    // Fetch current command and output when component loads
    this.fetchCurrentCommand();
    this.fetchCommandOutput();
    
    // Set up polling to check for command output
    this.pollingInterval = setInterval(() => {
      this.fetchCommandOutput();
    }, 5000); // Poll every 5 seconds
  },
  beforeDestroy() {
    // Clear the polling interval when component is destroyed
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
    }
  },
  methods: {
    async executeCommand() {
      if (!this.newCommand.trim()) return;

      const commandText = this.newCommand;
      this.newCommand = "";
      this.loading = true;

      try {
        // Check if this is a sudo command that might need a password
        const isSudoCommand = commandText.trim().startsWith('sudo ');
        
        // Add command to the list with appropriate status
        const newCmd = { 
          command: commandText, 
          response: "",
          status: isSudoCommand ? "interactive" : "pending",
          interactiveInput: "",
          inputPrompt: isSudoCommand ? "[sudo] password:" : "",
          inputType: isSudoCommand ? "password" : "text",
          partialResponse: ""
        };
        
        this.commands.push(newCmd);
        
        if (!isSudoCommand) {
          // Send command to backend
          await this.sendCommandToAgent(commandText);
          
          // Refresh data
          await this.fetchCurrentCommand();
        }
      } catch (error) {
        console.error("Error executing command:", error);
        
        // Update the command status to show error
        const pendingCommand = this.commands.find(cmd => 
          cmd.command === commandText && cmd.status === "pending");
          
        if (pendingCommand) {
          pendingCommand.status = "complete";
          pendingCommand.response = "Error: Failed to send command to server";
        }
      } finally {
        this.loading = false;
      }
    },
    async sendInteractiveInput(cmd) {
      if (!cmd.interactiveInput) return;
      
      const input = cmd.interactiveInput;
      cmd.interactiveInput = "";
      
      try {
        // For sudo commands, we'll send the command with the password
        if (cmd.command.startsWith('sudo ') && cmd.inputPrompt === "[sudo] password:") {
          // Store the password in the command object temporarily
          const password = input;
          
          // Update the status to pending
          cmd.status = "pending";
          
          // Send the command with the password
          await this.sendCommandToAgent(cmd.command, password);
          
          // Refresh data
          await this.fetchCurrentCommand();
        } else {
          // For other interactive inputs
          // Update partial response to show the input (except for passwords)
          if (cmd.inputType !== 'password') {
            cmd.partialResponse += `\n${input}`;
          }
          
          // Send the input to the agent
          await this.sendInteractiveInputToAgent(cmd.command, input);
        }
      } catch (error) {
        console.error("Error sending interactive input:", error);
        cmd.status = "complete";
        cmd.response = "Error: Failed to send input to server";
      }
    },
    async sendCommandToAgent(command, password = "") {
      // Send command to backend API
      const response = await fetch('http://localhost:8080/command', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          command: command,
          arguments: password ? password : "" // Pass password as arguments if provided
        })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      
      return await response.json();
    },
    async sendInteractiveInputToAgent(command, input) {
      // Send interactive input to backend API
      const response = await fetch('http://localhost:8080/interactive-input', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          command: command,
          input: input
        })
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      
      return await response.json();
    },
    async fetchCurrentCommand() {
      try {
        const response = await fetch('http://localhost:8080/command');
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.data && data.data.length > 0) {
          this.activeCommand = data.data[0];
        }
      } catch (error) {
        console.error("Error fetching current command:", error);
      }
    },
    async fetchCommandOutput() {
      try {
        const response = await fetch('http://localhost:8080/output');
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.data && data.data.length > 0) {
          const output = data.data[0];
          
          // Find any pending or interactive command that matches this output
          const pendingCommand = this.commands.find(cmd => 
            cmd.command === output.given_command && 
            (cmd.status === "pending" || cmd.status === "interactive"));
            
          if (pendingCommand) {
            // Update the command with the response
            pendingCommand.response = output.output;
            pendingCommand.status = "complete";
          } else {
            // If no pending command matches, add a new entry
            const existingCommand = this.commands.find(cmd => 
              cmd.command === output.given_command);
              
            if (!existingCommand) {
              this.commands.push({
                command: output.given_command,
                response: output.output,
                status: "complete"
              });
            } else if (existingCommand.status !== "pending" && existingCommand.status !== "interactive") {
              // Update existing command with response if it's not pending or interactive
              existingCommand.response = output.output;
              existingCommand.status = "complete";
            }
          }
        }
      } catch (error) {
        console.error("Error fetching command output:", error);
      }
    },
    goToFiles() {
      this.$router.push('/files');
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
.text-warning {
  color: #ffc107 !important;
}
.text-blue {
  color: #17a2b8 !important;
}
.interactive-input {
  display: flex;
  margin-top: 8px;
}
</style>
