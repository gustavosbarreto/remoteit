<template>
  <div>
    <v-container grid-list-xl fluid>
      <v-layout row wrap>
        <v-flex sm12>
          <h1>Devices</h1>
        </v-flex>
        <v-flex lg12>
          <v-data-table :headers="headers" :items="devices" class="elevation-1">
            <template v-slot:items="props">
              <td>{{ props.item.uid }}</td>
              <td>{{ props.item.public_key }}</td>
            </template>
          </v-data-table>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
import ApiService from "@/services/api";

export default {
  data() {
    return {
      headers: [
        {
          text: "UID",
          align: "left",
          sortable: false,
          value: "uid"
        },
        { text: "Public Key", sortable: false, value: "public_key" }
      ],
      devices: []
    };
  },

  async created() {
    this.devices = await ApiService.get("/devices").then(res => res.data);
  }
};
</script>
