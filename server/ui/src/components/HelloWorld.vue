<template>
  <v-container>
    <v-expansion-panel>
      <v-expansion-panel-content v-for="device in devices" :key="device.uid">
        <div slot="header">{{ device.uid }}</div>
        <v-card>
          <v-card-text>
            <code>{{ JSON.stringify(device.identity, null, 4) }}</code>
          </v-card-text>

          <v-card-actions>
            <v-layout align-center justify-end>
              <v-card-text class="grey--text darken-1">Ultima vez online: {{ device.last_seen }}</v-card-text>
              <v-btn flat color="orange">Open Terminal</v-btn>
            </v-layout>
          </v-card-actions>
        </v-card>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-container>
</template>

<script>
export default {
  data() {
    return {
      devices: []
    };
  },

  async mounted() {
    this.devices = await this.getDevices();
  },

  methods: {
    async getDevices() {
      return await this.$http.get("/api/devices").then(res => {
        return res.data;
      });
    }
  }
};
</script>

<style scoped>
code {
  width: 100%;
  padding: 20px;
}

code::before {
  content: "";
}
</style>
