<template>
  <v-container>
    <v-expansion-panel v-model="panel" popout>
      <v-expansion-panel-content v-for="(device, index) in devices" :key="device.uid">
        <div slot="header">
          <v-label>
            <v-icon>developer_board</v-icon>
            {{ device.uid }}
          </v-label>
        </div>
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

          <v-toolbar flat color="white">
            <v-toolbar-title>Real-time Log Watcher</v-toolbar-title>
          </v-toolbar>

          <v-data-table :headers="headers" :items="logs[index]" :hide-actions="true">
            <template slot="items" slot-scope="props">
              <td>
                <v-chip
                  :color="messageColor(props.item.level)"
                  text-color="white"
                  small
                >{{ props.item.level }}</v-chip>
              </td>
              <td>{{ props.item.message }}</td>
            </template>
          </v-data-table>
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
      devices: [],
      logs: {},
      headers: [
        {
          text: "Level",
          align: "left",
          sortable: false,
          value: "level"
        },
        {
          text: "Message",
          align: "left",
          sortable: false,
          value: "message"
        }
      ],
      panel: [],
      ws: null
    };
  },

  async mounted() {
    this.devices = await this.getDevices();
  },

  watch: {
    panel(index) {
      if (index !== null) {
        var device = this.devices[index];

        this.ws = new WebSocket(`ws://${location.host}/log/ws/${device.uid}`);

        var self = this;

        if (!this.logs[index]) {
          this.$set(this.logs, index, []);
        }

        this.ws.onmessage = function(e) {
          self.logs[index].push(JSON.parse(e.data));
        };
      } else {
        this.ws.close();
      }
    }
  },

  methods: {
    async getDevices() {
      return await this.$http.get("/api/devices").then(res => {
        return res.data;
      });
    },

    messageColor(level) {
      return {
        emerg: "red",
        alert: "orange",
        crit: "blue",
        err: "deep-orange",
        warning: "yellow",
        notice: "green",
        info: "blue",
        debug: "grey"
      }[level];
    },

    showLogs(device, index) {
      var ws = new WebSocket(`ws://${location.host}/log/ws/${device.uid}`);

      var self = this;

      if (!this.logs[index]) {
        this.$set(this.logs, index, []);
      }

      ws.onmessage = function(e) {
        self.$nextTick(() => {
          self.logs[index].push(JSON.parse(e.data));
        });
      };
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
