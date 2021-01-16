<template>
  <div>
    <h1>
      {{ recording.title }}
      <span v-if="recording.subtitle">~ {{ recording.subtitle }} ~</span>
    </h1>
    <audio
      :style="{ display: ready ? 'block' : 'none' }"
      ref="audioElement"
      :src="file"
      controls
      autoplay
    ></audio>
    <div class="info-wrapper">
      <span
        v-if="recording.description"
        class="description"
        v-html="recording.description"
      />
      <template v-if="recording.info">
        <hr v-if="recording.description" />
        <span v-html="recording.info" />
      </template>
      <template v-if="recording.url">
        <hr v-if="recording.info" />
        <a v-if="recording.url" :href="recording.url">Radio Link</a>
      </template>
    </div>
    <div v-if="!ready">
      The recording file is not ready. Current status is '{{ status }}'.<br />
      Please wait until the download will finish.
    </div>
  </div>
</template>

<script lang="ts">
import {
  defineComponent,
  ref,
  computed,
  onUnmounted,
  onMounted,
  reactive,
} from "vue";
import { useRoute } from "vue-router";
import { api, RecordingDetail } from "../api";

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
    const recording = reactive({
      title: "",
      subtitle: "",
      description: "",
      info: "",
      url: "",
    });
    api.getRecording(stationId, start).then((r) => {
      if (r === null) {
        return;
      }
      status.value = r.status.status;

      recording.title = r.recording.title;
      recording.description = r.recording.description;
      recording.subtitle = r.recording.subtitle;
      recording.info = r.recording.info;
      recording.url = r.recording.url;
    });

    const playerStateRepository = new PlayerStateRepository();
    playerStateRepository.load();
    const audioElement = ref<HTMLAudioElement | null>(null);

    let dispose: any = null;
    onMounted(() => {
      const a = audioElement.value;
      if (a === null) {
        return;
      }
      // Restore previous position
      a.currentTime = playerStateRepository.pos(stationId, start);
      a.play();

      // Periodically store playback position
      const h = setInterval(() => {
        const pos = a.currentTime;
        if (pos !== undefined) {
          playerStateRepository.store(stationId, start, pos);
        }
      }, 1000);
      dispose = () => {
        clearInterval(h);
        const pos = a.currentTime;
        if (pos !== undefined) {
          playerStateRepository.store(stationId, start, pos);
        }
      };
    });
    onUnmounted(() => {
      if (dispose !== null) {
        dispose();
      }
    });

    return {
      file: url,
      status,
      ready: computed(() => {
        return status.value === "READY";
      }),
      audioElement,
      recording,
    };
  },
});
</script>
<style scoped>
audio {
  width: 600px;
}
.info-wrapper {
  border: solid 3px #eee;
  padding: 1rem;
  display: inline-block;
  margin: 1rem;
  max-width: 600px;
  font-size: 0.8rem;
}
</style>
