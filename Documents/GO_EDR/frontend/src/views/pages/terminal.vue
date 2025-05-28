<template>
    <div class="terminal-container">
      <div class="noise"></div>
      <div class="terminal-header">
        <div>
          <span class="status-indicator" :class="{ blink: isConnected }"></span>
          <span>SECURE SHELL v3.4.21</span>
        </div>
        <div>
          <span>{{ currentDateTime }}</span>
        </div>
      </div>
      
      <div class="terminal">
        <div class="terminal-output" ref="output">
          <div class="response success">
            ██╗  ██╗ █████╗  ██████╗██╗  ██╗███████╗██████╗ <br>
            ██║  ██║██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗<br>
            ███████║███████║██║     █████╔╝ █████╗  ██████╔╝<br>
            ██╔══██║██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗<br>
            ██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║<br>
            ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝<br>
            <br>
            T E R M I N A L
          </div>
          <div class="response">
            Initializing secure connection to server...<br>
            Connection established. Authentication required.
          </div>
          <div class="response success">
            Authentication successful. Terminal ready.<br>
            Type 'help' for available commands.
          </div>
          
          <div v-for="(cmd, index) in commands" :key="index">
            <div class="command">
              <span class="prompt">root@server:~#</span> {{ cmd.command }}
            </div>
            <div v-if="cmd.status === 'pending'" class="response warning">
              <span class="spinner"></span>
              Command issued, waiting for response...
            </div>
            <div v-else-if="cmd.status === 'interactive'" class="response warning">
              <pre v-if="cmd.partialResponse" class="response-text">{{ cmd.partialResponse }}</pre>
              <div class="interactive-input">
                <input 
                  type="text" 
                  v-model="cmd.interactiveInput" 
                  :placeholder="cmd.inputPrompt || 'Enter input...'" 
                  :type="cmd.inputType || 'text'"
                  @keydown.enter="sendInteractiveInput(cmd)"
                  class="interactive-input-field"
                >
                <button class="button ml-2" @click="sendInteractiveInput(cmd)">
                  Send
                </button>
              </div>
            </div>
            <div v-else class="response" :class="cmd.class">
              <pre class="response-text">{{ cmd.response }}</pre>
              <div v-if="cmd.command === 'send' && cmd.status === 'complete'" class="mt-2">
                <button class="button success-button" @click="goToFiles">
                  View Uploaded Files
                </button>
              </div>
            </div>
          </div>
        </div>
        
        <div class="input-line">
          <span class="prompt">root@server:~#</span>
          <input 
            type="text" 
            v-model="commandInput" 
            @keydown.enter="executeCommand" 
            ref="terminalInput"
            :disabled="!isConnected || loading || hasInteractiveCommand"
          >
        </div>
      </div>
      
      <div class="status-bar">
        <div>
          <span>STATUS: </span>
          <span>{{ connectionStatus }}</span>
          <span v-if="loading" class="loading-indicator">PROCESSING</span>
        </div>
        <div>
          <button class="button" @click="clearTerminal">CLEAR</button>
          <button class="button" @click="disconnect">DISCONNECT</button>
        </div>
      </div>
    </div>
  </template>
  
  <script>
  export default {
    data() {
      return {
        commandInput: '',
        commands: [],
        isConnected: true,
        connectionStatus: 'CONNECTED',
        currentDateTime: '',
        dateTimeInterval: null,
        loading: false,
        activeCommand: null,
        pollingInterval: null
      }
    },
    computed: {
      hasInteractiveCommand() {
        return this.commands.some(cmd => cmd.status === 'interactive');
      }
    },
    mounted() {
      this.updateDateTime()
      this.dateTimeInterval = setInterval(this.updateDateTime, 1000)
      this.$refs.terminalInput.focus()
      
      // Add event listener to focus on terminal input when clicking the terminal
      this.$refs.output.parentNode.addEventListener('click', this.focusInput)
      
      // Set up polling to check for command output
      this.fetchCurrentCommand()
      this.fetchCommandOutput()
      this.pollingInterval = setInterval(() => {
        this.fetchCommandOutput()
      }, 5000) // Poll every 5 seconds
    },
    beforeDestroy() {
      clearInterval(this.dateTimeInterval)
      clearInterval(this.pollingInterval)
      this.$refs.output.parentNode.removeEventListener('click', this.focusInput)
    },
    methods: {
      updateDateTime() {
        const now = new Date()
        this.currentDateTime = now.toLocaleString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          year: 'numeric',
          month: '2-digit',
          day: '2-digit'
        })
      },
      async executeCommand() {
        const command = this.commandInput.trim()
        if (!command || !this.isConnected || this.loading || this.hasInteractiveCommand) return
        
        this.commandInput = ''
        this.loading = true
        
        try {
          // Check if this is a sudo command that might need a password
          const isSudoCommand = command.startsWith('sudo ')
          
          // Add command to history
          const newCmd = { 
            command: command, 
            response: "",
            status: isSudoCommand ? "interactive" : "pending",
            class: '',
            interactiveInput: "",
            inputPrompt: isSudoCommand ? "[sudo] password:" : "",
            inputType: isSudoCommand ? "password" : "text",
            partialResponse: ""
          }
          
          this.commands.push(newCmd)
          
          // Scroll to bottom
          this.$nextTick(() => {
            this.$refs.output.scrollTop = this.$refs.output.scrollHeight
          })
          
          if (!isSudoCommand) {
            // Send command to backend
            await this.sendCommandToAgent(command)
            
            // Refresh data
            await this.fetchCurrentCommand()
          }
        } catch (error) {
          console.error("Error executing command:", error)
          
          // Update the command status to show error
          const pendingCommand = this.commands.find(cmd => 
            cmd.command === command && cmd.status === "pending")
            
          if (pendingCommand) {
            pendingCommand.status = "complete"
            pendingCommand.response = "Error: Failed to send command to server"
            pendingCommand.class = "error"
          }
        } finally {
          this.loading = false
        }
      },
      async sendInteractiveInput(cmd) {
        if (!cmd.interactiveInput) return
        
        const input = cmd.interactiveInput
        cmd.interactiveInput = ""
        
        try {
          
          if (cmd.command.startsWith('sudo ') && cmd.inputPrompt === "[sudo] password:") {
            // Store the password in the command object temporarily
            const password = input
            
            // Update the status to pending
            cmd.status = "pending"
            
            // Send the command with the password
            await this.sendCommandToAgent(cmd.command, password)
            
            // Refresh data
            await this.fetchCurrentCommand()
          } else {
            // For other interactive inputs
            // Update partial response to show the input (except for passwords)
            if (cmd.inputType !== 'password') {
              cmd.partialResponse += `\n${input}`
            }
            
            // Send the input to the agent
            await this.sendInteractiveInputToAgent(cmd.command, input)
          }
        } catch (error) {
          console.error("Error sending interactive input:", error)
          cmd.status = "complete"
          cmd.response = "Error: Failed to send input to server"
          cmd.class = "error"
        }
      },
      async sendCommandToAgent(command, password = "") {
        // Send command to backend API
        const response = await fetch('http://localhost:8080/command', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          credentials: 'include', // Include credentials for cross-origin requests
          body: JSON.stringify({
            command: command,
            arguments: password ? password : "" // Pass password as arguments if provided
          })
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`)
        }
        
        return await response.json()
      },
      async sendInteractiveInputToAgent(command, input) {
        // Send interactive input to backend API
        const response = await fetch('http://localhost:8080/interactive-input', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          credentials: 'include', // Include credentials for cross-origin requests
          body: JSON.stringify({
            command: command,
            input: input
          })
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`)
        }
        
        return await response.json()
      },
      async fetchCurrentCommand() {
        try {
          const response = await fetch('http://localhost:8080/command', {
            credentials: 'include', // Include credentials for cross-origin requests
            headers: {
              'Accept': 'application/json'
            }
          })
          
          if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`)
          }
          
          const data = await response.json()
          if (data.data && data.data.length > 0) {
            this.activeCommand = data.data[0]
          }
        } catch (error) {
          console.error("Error fetching current command:", error)
        }
      },
      async fetchCommandOutput() {
        try {
          const response = await fetch('http://localhost:8080/output', {
            credentials: 'include', // Include credentials for cross-origin requests
            headers: {
              'Accept': 'application/json'
            }
          })
          
          if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`)
          }
          
          const data = await response.json()
          if (data.data && data.data.length > 0) {
            const output = data.data[0]
            
            // Find any pending or interactive command that matches this output
            const pendingCommand = this.commands.find(cmd => 
              cmd.command === output.given_command && 
              (cmd.status === "pending" || cmd.status === "interactive"))
              
            if (pendingCommand) {
              // Update the command with the response
              pendingCommand.response = output.output
              pendingCommand.status = "complete"
              pendingCommand.class = "success"
            } else {
              // If no pending command matches, add a new entry
              const existingCommand = this.commands.find(cmd => 
                cmd.command === output.given_command)
                
              if (!existingCommand) {
                this.commands.push({
                  command: output.given_command,
                  response: output.output,
                  status: "complete",
                  class: "success"
                })
              } else if (existingCommand.status !== "pending" && existingCommand.status !== "interactive") {
                // Update existing command with response if it's not pending or interactive
                existingCommand.response = output.output
                existingCommand.status = "complete"
                existingCommand.class = "success"
              }
            }
            
            // Scroll to bottom after updating content
            this.$nextTick(() => {
              this.$refs.output.scrollTop = this.$refs.output.scrollHeight
            })
          }
        } catch (error) {
          console.error("Error fetching command output:", error)
        }
      },
      clearTerminal() {
        this.commands = []
      },
      disconnect() {
        this.isConnected = false
        this.connectionStatus = 'DISCONNECTED'
        
        // Add disconnection message
        this.commands.push({
          command: 'disconnect',
          response: 'Connection terminated.',
          status: 'complete',
          class: 'error'
        })
        
        // Scroll to bottom after adding message
        this.$nextTick(() => {
          this.$refs.output.scrollTop = this.$refs.output.scrollHeight
        })
      },
      focusInput() {
        if (this.isConnected && !this.hasInteractiveCommand) {
          this.$refs.terminalInput.focus()
        }
      },
      goToFiles() {
        // Navigate to files view (assume using vue-router)
        // this.$router.push('/files')
        alert('Navigate to files view')
      }
    }
  }
  </script>
  
  <style scoped>
  :root {
    --bg-color: #0a0a0a;
    --terminal-color: #0d1117;
    --text-color: #00ff41;
    --accent-color: #0f0;
    --error-color: #ff073a;
    --header-color: #00ccff;
    --warning-color: #ffc107;
    --success-color: #0f0;
    --font-family: 'Courier New', monospace;
  }
  
  .terminal-container {
    background-color: var(--bg-color);
    color: var(--text-color);
    font-family: var(--font-family);
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
  }
  
  .terminal-header {
    border-bottom: 1px solid var(--header-color);
    padding: 15px 20px;
    color: var(--header-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .terminal {
    background-color: var(--terminal-color);
    border-radius: 5px;
    border: 1px solid var(--accent-color);
    box-shadow: 0 0 10px var(--accent-color);
    flex: 1;
    margin: 15px;
    padding: 15px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  
  .terminal-output {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    margin-bottom: 10px;
  }
  
  .input-line {
    display: flex;
    width: 100%;
  }
  
  .prompt {
    color: var(--accent-color);
    margin-right: 10px;
  }
  
  input {
    background: transparent;
    border: none;
    color: var(--text-color);
    font-family: var(--font-family);
    font-size: 1em;
    width: 100%;
    outline: none;
  }
  
  .command {
    margin-bottom: 10px;
    word-break: break-all;
  }
  
  .response {
    margin-bottom: 15px;
    color: #aaa;
    word-break: break-all;
  }
  
  .response-text {
    font-family: var(--font-family);
    margin: 0;
    color: var(--text-color);
  }
  
  .success {
    color: var(--success-color);
  }
  
  .error {
    color: var(--error-color);
  }
  
  .warning {
    color: var(--warning-color);
  }
  
  .blink {
    animation: blink 1s infinite;
  }
  
  @keyframes blink {
    0%, 49% {
      opacity: 1;
    }
    50%, 100% {
      opacity: 0;
    }
  }
  
  .noise {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    opacity: 0.05;
    background-image: url('data:image/png;base64,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');
  }
  
  .status-indicator {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: var(--accent-color);
    display: inline-block;
    margin-right: 10px;
  }
  
  /* Status Bar at bottom */
  .status-bar {
    display: flex;
    justify-content: space-between;
    padding: 10px 15px;
    background-color: var(--terminal-color);
    border-top: 1px solid var(--accent-color);
    font-size: 0.8em;
  }
  
  .button {
    background-color: transparent;
    color: var(--accent-color);
    border: 1px solid var(--accent-color);
    border-radius: 3px;
    padding: 2px 8px;
    font-family: var(--font-family);
    cursor: pointer;
    transition: all 0.3s;
  }
  
  .button:hover {
    background-color: var(--accent-color);
    color: var(--bg-color);
  }
  
  .success-button {
    background-color: transparent;
    color: var(--success-color);
    border: 1px solid var(--success-color);
  }
  
  .success-button:hover {
    background-color: var(--success-color);
    color: var(--bg-color);
  }
  
  .interactive-input {
    display: flex;
    margin-top: 8px;
  }
  
  .interactive-input-field {
    flex: 1;
    background-color: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--accent-color);
    padding: 5px;
  }
  
  .ml-2 {
    margin-left: 8px;
  }
  
  .mt-2 {
    margin-top: 8px;
  }
  
  .loading-indicator {
    color: var(--warning-color);
    margin-left: 10px;
  }
  
  .spinner {
    display: inline-block;
    width: 10px;
    height: 10px;
    border: 2px solid transparent;
    border-top-color: var(--warning-color);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 8px;
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  </style>