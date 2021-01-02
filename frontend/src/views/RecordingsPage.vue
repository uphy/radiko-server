<template>
  <div>
    <Recordings :recordings="state.recordings" />
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive } from "vue";
import Recordings from "@/components/Recordings.vue"; // @ is an alias to /src
import { api, Recording } from "../api/";

interface State {
  recordings: Recording[];
}

export default defineComponent({
  name: "RecordingsPage",
  setup() {
    const state = reactive<State>({
      recordings: [],
    });
    api.getRecordings().then((recordings) => {
      state.recordings = recordings;
    });

    return { state };
  },
  components: {
    Recordings,
  },
});
</script>
