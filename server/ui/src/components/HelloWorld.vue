<template>
  <v-container>{{ devices }}</v-container>
</template>

<script>
export default {
  data: () => {
    devices: []
  },
  mounted() {
    this.devices = await this.$http.get('/devices');
  }
};
</script>

<style scoped>
body {
  color: red;
}
</style>
