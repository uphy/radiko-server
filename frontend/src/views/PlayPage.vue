<template>
  <audio :src="file" controls autoplay></audio>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useRoute } from "vue-router";
import { api } from "../api";

export default defineComponent({
  name: "PlayPage",
  setup() {
    const route = useRoute();
    const stationId: string = route.params.stationId as string;
    const start: string = route.params.start as string;
    const url = api.getMp3Url(stationId, start);
    return {
      file: url,
    };
  },
});
</script>
