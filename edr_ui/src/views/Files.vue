<template>
  <div>
    <b-container fluid class="mt-4">
      <b-row>
        <b-col>
          <b-card class="shadow">
            <b-card-header>
              <h3 class="mb-0">Uploaded Files</h3>
            </b-card-header>
            <b-card-body>
              <b-table 
                striped 
                hover 
                :items="files" 
                :fields="fields"
                responsive
                sort-by="timestamp"
                sort-desc
              >
                <template #cell(filename)="data">
                  <a :href="getFileUrl(data.item)" target="_blank">{{ data.item.filename }}</a>
                </template>
                <template #cell(actions)="data">
                  <b-button 
                    size="sm" 
                    variant="primary" 
                    @click="viewFile(data.item)"
                    class="mr-2"
                  >
                    View
                  </b-button>
                  <b-button 
                    size="sm" 
                    variant="success" 
                    @click="downloadFile(data.item)"
                  >
                    Download
                  </b-button>
                </template>
              </b-table>
            </b-card-body>
          </b-card>
        </b-col>
      </b-row>
    </b-container>

    <!-- File Preview Modal -->
    <b-modal 
      v-model="showPreview" 
      size="lg" 
      title="File Preview" 
      ok-only
      ok-title="Close"
    >
      <div v-if="selectedFile">
        <div v-if="isImage(selectedFile)">
          <img :src="getFileUrl(selectedFile)" class="img-fluid" alt="File preview" />
        </div>
        <div v-else-if="isText(selectedFile)">
          <pre class="file-content">{{ fileContent }}</pre>
        </div>
        <div v-else>
          <p>Preview not available for this file type.</p>
          <b-button variant="primary" @click="downloadFile(selectedFile)">Download File</b-button>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script>
export default {
  data() {
    return {
      files: [],
      fields: [
        { key: 'filename', label: 'Filename', sortable: true },
        { key: 'agent', label: 'Agent', sortable: true },
        { key: 'timestamp', label: 'Upload Time', sortable: true },
        { key: 'actions', label: 'Actions' }
      ],
      showPreview: false,
      selectedFile: null,
      fileContent: ''
    };
  },
  created() {
    this.fetchFiles();
    
    // Refresh files list every 30 seconds
    setInterval(() => {
      this.fetchFiles();
    }, 30000);
  },
  methods: {
    async fetchFiles() {
      try {
        const response = await fetch('http://localhost:8080/files');
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        const data = await response.json();
        if (data.data) {
          // Add timestamp from filename if not provided by API
          this.files = data.data.map(file => {
            // Extract timestamp from filename if it contains one
            if (!file.timestamp && file.filename) {
              const match = file.filename.match(/\d{8}-\d{6}/);
              if (match) {
                const timestamp = match[0];
                const formattedDate = `${timestamp.substring(0, 4)}-${timestamp.substring(4, 6)}-${timestamp.substring(6, 8)} ${timestamp.substring(9, 11)}:${timestamp.substring(11, 13)}:${timestamp.substring(13, 15)}`;
                file.timestamp = formattedDate;
              }
            }
            return file;
          });
        }
      } catch (error) {
        console.error('Error fetching files:', error);
      }
    },
    getFileUrl(file) {
      return `http://localhost:8080${file.file_url}`;
    },
    async viewFile(file) {
      this.selectedFile = file;
      
      if (this.isText(file)) {
        try {
          const response = await fetch(this.getFileUrl(file));
          if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
          }
          
          this.fileContent = await response.text();
        } catch (error) {
          console.error('Error fetching file content:', error);
          this.fileContent = 'Error loading file content.';
        }
      }
      
      this.showPreview = true;
    },
    downloadFile(file) {
      const link = document.createElement('a');
      link.href = this.getFileUrl(file);
      link.download = file.filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    isImage(file) {
      const ext = this.getFileExtension(file.filename).toLowerCase();
      return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp'].includes(ext);
    },
    isText(file) {
      const ext = this.getFileExtension(file.filename).toLowerCase();
      return ['txt', 'log', 'json', 'xml', 'html', 'css', 'js', 'md', 'csv', 'ini', 'conf', 'sh', 'py', 'go', 'c', 'cpp', 'h', 'java'].includes(ext);
    },
    getFileExtension(filename) {
      return filename.split('.').pop() || '';
    }
  }
};
</script>

<style>
.file-content {
  max-height: 500px;
  overflow-y: auto;
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 5px;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style> 