<template>
  <div class="login-container d-flex align-items-center justify-content-center">
    <b-card class="login-card shadow-lg p-4">
      <!-- Replaced Logo with "AWK" -->
      <b-card-title class="text-center text-white text-uppercase font-weight-bold">
        AWK
      </b-card-title>

      <b-alert v-if="errorMessage" variant="danger" show>{{ errorMessage }}</b-alert>

      <b-form @submit.prevent="login">
        <b-form-group label="Username" label-class="text-white">
          <b-form-input v-model="username" required></b-form-input>
        </b-form-group>

        <b-form-group label="Password" label-class="text-white">
          <b-form-input v-model="password" type="password" required></b-form-input>
        </b-form-group>

        <b-button type="submit" variant="primary" block :disabled="loading">
          <b-spinner v-if="loading" small></b-spinner>
          <span v-else>Login</span>
        </b-button>
      </b-form>
    </b-card>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: "",
      password: "",
      errorMessage: "",
      loading: false
    };
  },
  methods: {
    async login() {
      if (!this.username || !this.password) {
        this.errorMessage = "Username and password are required.";
        return;
      }

      this.loading = true;
      this.errorMessage = "";

      try {
        // Replace with actual API call
        const response = await this.fakeLoginAPI(this.username, this.password);
        if (response.success) {
          this.$router.push("/dashboard");
        } else {
          this.errorMessage = "Invalid username or password.";
        }
      } catch (error) {
        this.errorMessage = "Login failed. Please try again.";
      } finally {
        this.loading = false;
      }
    },
    async fakeLoginAPI(username, password) {
      return new Promise((resolve) => {
        setTimeout(() => {
          resolve({ success: username === "admin" && password === "password123" });
        }, 1000);
      });
    }
  }
};
</script>

<style>
.login-container {
  height: 100vh;
  background: linear-gradient(45deg, #233554, #0a192f);
}

.login-card {
  width: 400px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 12px;
}
</style>
