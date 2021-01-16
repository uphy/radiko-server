<template>
  <div>
    URL:
    <input type="text" v-model="state.url" />
    <button @click="record" :disabled="state.downloading">Download</button>
    <div v-if="state.status.length > 0">
      Status: {{ state.status }}<br />
      Progress: {{ state.downloadProgress * 100 }} %
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive } from "vue";
import { api } from "../api";

export default defineComponent({
  name: "RecordPage",
  setup() {
    const state = reactive({
      url: "",
      status: "",
      downloadProgress: 0,
      downloading: false,
    });
    return {
      state,
      async record() {
        // https://radiko.jp/#!/ts/FMT/20201227230000
        const regexpUrl = /.*?\/ts\/(?<stationId>\w+)\/(?<start>\w+)/gm;
        const match = regexpUrl.exec(state.url);
        if (match === undefined || match === null) {
          return;
        }
        const stationId = match.groups?.stationId;
        const start = match.groups?.start;
        if (stationId !== undefined && start !== undefined) {
          state.url = "";
          state.downloading = true;
          api.record(stationId, start);

          /* eslint-disable */
          let handle = setInterval(async () => {
            const status = await api.getStatus(stationId, start);
            if (status === null) {
              return;
            }
            if (status.status !== "DOWNLOADING") {
              clearInterval(handle);
              state.status = "";
              state.downloadProgress = 0;
              state.downloading = false;
            } else {
              state.status = status.status;
              state.downloadProgress = status.downloadProgress;
            }
          }, 1000);
        }
      },
    };
  },
});
</script>
