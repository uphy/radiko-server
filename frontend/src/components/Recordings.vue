<template>
  <div>
    <div
      class="recordings"
      v-for="(recordings, i) in recordingsGroupByDate"
      :key="i"
    >
      <div class="recordings-date">{{ recordings.date.toString() }}</div>
      <div
        class="recording"
        v-for="(recording, i) in recordings.recordings"
        :key="i"
      >
        <div class="recording-title">
          <router-link :to="routerLink(recording.recording)">
            <span :style="{color: recording.rate === 100 ? '#d489fb': recording.rate > 0 ? '#8f00d8' : '#333'}">{{ recording.recording.title }}</span>
            <i v-if="recording.rate === 100" class="bi bi-check"></i>
            <i v-else-if="recording.rate > 0" class="bi bi-music-note"></i>
          </router-link>
        </div>
        <div class="recording-date">
          {{ recording.recording.start.toLocaleString() }} -
          {{ recording.recording.end.toLocaleString() }}
          <span class="octicons octicons-check"></span><span v-if="recording.rate > 0">({{ recording.rate }} % played)</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { Recording } from "../api";
import { playerStateRepository } from "../repositories";

type Props = {
  recordings: Recording[];
};

class RecordingsByDate {
  constructor(public date: MonthDate, public recordings: RecordingWrapper[]) {}
}

class RecordingWrapper {
  constructor(public recording: Recording, public rate: number) {}
}

class MonthDate {
  constructor(public year: number, public month: number, public date: number) {}
  equals(other: MonthDate): boolean {
    return (
      this.year === other.year &&
      this.month === other.month &&
      this.date === other.date
    );
  }
  toString(): string {
    return `${this.year}/${this.month}/${this.date}`;
  }
}

export default defineComponent({
  name: "Recordings",
  props: {
    recordings: {
      type: Array,
      default: () => [],
    },
  },
  setup<Props>(props: any) {
    const p = (s: string, len: number): string => {
      for (let i = s.length; i < len; i++) {
        s = "0" + s;
      }
      return s;
    };
    const dateToString = (date: Date) => {
      return `${date.getFullYear()}${p((date.getMonth() + 1).toString(), 2)}${p(
        date.getDate().toString(),
        2
      )}${p(date.getHours().toString(), 2)}${p(
        date.getMinutes().toString(),
        2
      )}${p(date.getSeconds().toString(), 2)}`;
    };
    const getPlayedRate = (recording: Recording): number => {
      const all = recording.end.getTime() - recording.start.getTime();
      if (all === 0) {
        return 0;
      }
      const rate =
        playerStateRepository.pos(
          recording.stationId,
          dateToString(recording.start)
        ) /
        (all / 1000);
      return Math.floor(rate * 100 + 0.5);
    };
    return {
      routerLink(recording: any): string {
        const start: Date = recording.start;
        return `/recordings/${recording.stationId}/${dateToString(start)}`;
      },
      recordingsGroupByDate: computed(() => {
        let result: RecordingsByDate[] = [];
        let current: RecordingsByDate | null = null;
        const sorted: Recording[] = [...props.recordings];
        sorted.sort((a: any, b: any) => {
          return b.start - a.start;
        });

        sorted.forEach((recording: any) => {
          const date = new MonthDate(
            recording.start.getFullYear(),
            recording.start.getMonth() + 1,
            recording.start.getDate()
          );
          if (current === null || !current.date.equals(date)) {
            current = new RecordingsByDate(date, []);
            result.push(current);
          }
          current.recordings.push(
            new RecordingWrapper(recording, getPlayedRate(recording))
          );
        });
        return result;
      }),
    };
  },
});
</script>

<style scoped>
.recordings {
  background-color: #e6c3f8;
  margin-bottom: 0.5rem;
}
.recordings-date {
  font-size: 1.5rem;
  display: block;
  padding: 0 0.5rem;
  margin: 0;
  background-color: #8f00d8;
  color: #dbdbdb;
}
.recording {
  padding: 0.5rem;
}
.recording-title {
  font-size: 1rem;
}
.recording-date {
  font-size: 0.5rem;
}
</style>
