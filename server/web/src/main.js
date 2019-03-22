import Vue from "vue";
import "./plugins/vuetify";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import ApiService from "@/services/api";

Vue.config.productionTip = false;

ApiService.init("/api");

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount("#app");
