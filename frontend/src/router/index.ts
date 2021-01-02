import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import RecordingsPage from "../views/RecordingsPage.vue";
import RecordPage from "../views/RecordPage.vue";
import PlayPage from "../views/PlayPage.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "Recordings",
    component: RecordingsPage,
  },
  {
    path: "/record",
    name: "Record",
    component: RecordPage,
  },
  {
    path: "/recordings/:stationId/:start",
    name: "Play",
    component: PlayPage,
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router;
