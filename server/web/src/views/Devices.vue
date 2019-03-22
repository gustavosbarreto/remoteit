<template>
  <div>
    <h1>Devices</h1>
  </div>
</template>

<script>
import ApiService from "@/services/api";

export default {
  created() {
    ApiService.get("/devices");
  }
};
</script>
