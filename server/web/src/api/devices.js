import ApiService from "@/services/api";

const DevicesAPI = {
  getAll() {
    return ApiService.get("/devices");
  }
};

export default DevicesAPI;
