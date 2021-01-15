<template>
  <div>
    <audio v-if="ready" ref="audioElement" :src="file" controls autoplay></audio>
    <span v-else>
      The recording file is not ready. Current status is '{{ status }}'.<br />
      Please wait until the download will finish.
    </span>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { api } from "../api";

class PlayerState {
  constructor(public pos: number, public lastPlay: Date) {}
}

class PlayerStateRepository {
  data: any;
  load() {
    const p = localStorage.getItem("player");
    if (p !== null) {
      this.data = JSON.parse(p);
    } else {
      this.data = {};
    }
  }
  store(stationId: string, start: string, pos: number) {
    this.data[`${stationId}-${start}`] = new PlayerState(pos, new Date());
    localStorage.setItem("player", JSON.stringify(this.data));
  }
  pos(stationId: string, start: string): number {
    const state = this.data[`${stationId}-${start}`];
    if (state !== undefined) {
      return state.pos;
    }
    return 0;
  }
}

export default defineComponent({
  name: "PlayPage",
  setup() {
    const route = useRoute();
    const stationId: string = route.params.stationId as string;
    const start: string = route.params.start as string;
    const url = api.getMp3Url(stationId, start);
    const status = ref("");
    api.getStatus(stationId, start).then((s) => {
      status.value = s.status;
    });

    const playerStateRepository = new PlayerStateRepository();
    playerStateRepository.load();
    const audioElement = ref<HTMLAudioElement | null>(null);
    let first = true;
    const h = setInterval(() => {
      const a = audioElement.value;
      if (a === null) {
        return;
      }
      if (first) {
        a.currentTime = playerStateRepository.pos(stationId, start);
        first = false;
        return;
      }

      const pos = a.currentTime;
      if (pos !== undefined) {
        playerStateRepository.store(stationId, start, pos);
      }
    }, 1000);
    onUnmounted(() => {
      clearInterval(h);
    });

    return {
      file: url,
      status,
      ready: computed(() => {
        return status.value === "READY";
      }),
      audioElement,
    };
  },
});
</script>
