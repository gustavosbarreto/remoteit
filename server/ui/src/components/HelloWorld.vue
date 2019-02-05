<template>
  <v-container>
    <v-expansion-panel>
      <v-expansion-panel-content v-for="(device, index) in devices" :key="device.uid">
        <div slot="header">{{ device.uid }}</div>
        <v-card>
          <v-card-text>
            <code>{{ JSON.stringify(device.identity, null, 4) }}</code>
          </v-card-text>

          <v-card-actions>
            <v-layout align-center justify-end>
              <v-card-text class="grey--text darken-1">Ultima vez online: {{ device.last_seen }}</v-card-text>
              <v-btn flat color="orange" @click="openTerminal(device, index)">Open Terminal</v-btn>
            </v-layout>
          </v-card-actions>

          <div class="terminal" ref="terminal"></div>
        </v-card>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-container>
</template>

<script>
import { Terminal } from "xterm";
import * as fit from "xterm/lib/addons/fit/fit";
import "xterm/dist/xterm.css";

Terminal.applyAddon(fit);

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
    },

    openTerminal(device, index) {
      const xterm = new Terminal({
        cursorBlink: true,
        fontFamily: "monospace"
      });

      xterm.open(this.$refs.terminal[index]);
      xterm.focus();
      xterm.fit();

      const params = Object.entries({
        user: `gustavo@${device.uid}`,
        cols: xterm.cols,
        rows: xterm.rows
      })
        .map(([k, v]) => {
          return `${k}=${v}`;
        })
        .join("&");

      var ws = new WebSocket(`ws://${location.host}/term/ws?${params}`);

      ws.onmessage = function(e) {
        xterm.write(e.data);
      };

      xterm.on("data", data => {
        ws.send(data);
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

.terminal {
  margin: 20px;
}
</style>
