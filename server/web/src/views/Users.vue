<template>
  <div>
    <v-container grid-list-xl fluid>
      <v-layout row wrap>
        <v-flex sm12>
          <v-layout row>
            <v-flex grow>
              <h1>Users</h1>
            </v-flex>
            <v-flex shrink>
              <v-btn fab small color="primary">
                <v-icon>add</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-flex>
        <v-flex lg12>
          <v-data-table :headers="headers" :items="users" :hide-actions="true" class="elevation-1">
            <template v-slot:items="props">
              <td>{{ props.item.name }}</td>
              <td>{{ props.item.username }}</td>
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
          text: "Name",
          align: "left",
          sortable: false,
          value: "name"
        },
        { text: "Username", sortable: false, value: "username" }
      ],
      users: []
    };
  },

  async created() {
    this.users = await ApiService.get("/users").then(res => res.data);
  }
};
</script>
