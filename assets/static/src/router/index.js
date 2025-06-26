import FileView from "@/views/FileView.vue";
import PasteView from "@/views/PasteView.vue";
import UploadView from "@/views/UploadView.vue";
import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(window.config?.site_path),
  routes: [
    {
      path: "/",
      name: "Upload",
      component: UploadView,
      meta: {
        navigation: true,
      },
    },
    {
      path: "/paste",
      name: "Paste",
      component: PasteView,
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
      component: FileView,
      props: true,
    },
  ],
});

router.beforeEach((to) => {
  if (to.name !== "File") {
    document.title = to.name + " · " + window.config.site_name;
  }
});

export default router;
