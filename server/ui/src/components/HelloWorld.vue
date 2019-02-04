<template>
  <v-container>ae</v-container>
</template>

<script>
export default {
  data() {
    return {
      devices: []
    };
  },
  async mounted() {
    this.devices = await this.$http.get("/api/devices");
  }
};
</script>

<style scoped>
body {
  color: red;
}
</style>
