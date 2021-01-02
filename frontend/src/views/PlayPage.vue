<template>
  <div>Play <Player :options="options" /></div>
</template>

<script lang="ts">
import { defineComponent, reactive } from "vue";
import { useRoute } from "vue-router";
import Player from "../components/Player.vue";
import { api } from "../api";
import "video.js/dist/video-js.css";

export default defineComponent({
  name: "PlayPage",
  setup() {
    const route = useRoute();
    console.log(route.params);
    const stationId: string = route.params.stationId as string;
    const start: string = route.params.start as string;
    const url = api.getPlaylistUrl(stationId, start);
    console.log(url);
    return {
      options: {
        autoplay: true,
        controls: true,
        height: 100,
        sources: [
          {
            src: url,
            type: "application/x-mpegURL",
          },
        ],
      },
    };
  },
  components: {
    Player,
  },
});
</script>
