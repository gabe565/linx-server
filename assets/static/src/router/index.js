import { useConfigStore } from "@/stores/config.js";
import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Upload",
      component: () => import("../views/UploadView.vue"),
      meta: {
        navigation: true,
      },
    },
    {
      path: "/paste",
      name: "Paste",
      component: () => import("../views/PasteView.vue"),
      meta: {
        navigation: true,
      },
    },
    {
      path: "/api",
      name: "API",
      component: () => import("../views/APIView.vue"),
      meta: {
        navigation: true,
      },
    },
    {
      path: "/:filename(.*)",
      name: "File",
      component: () => import("../views/FileView.vue"),
      props: true,
    },
  ],
});

router.beforeEach((to) => {
  if (to.name !== "File") {
    document.title = to.name + " · " + useConfigStore().site.site_name;
  }
});

export default router;
