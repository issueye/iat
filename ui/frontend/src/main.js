import { createApp } from "vue";
import { createPinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import naive from "naive-ui";
import ElementPlusX from "vue-element-plus-x";
import App from "./App.vue";
import router from "./router";
import "./style.css";
import hljs from "highlight.js";
import "highlight.js/styles/atom-one-dark.css";
import "@/assets/css/index.css";

const app = createApp(App);

// Register global components
// app.component("XMarkdown", XMarkdown);

// Register highlight directive
app.directive("highlight", {
  mounted(el) {
    const blocks = el.querySelectorAll("code");
    blocks.forEach((block) => {
      hljs.highlightElement(block);
    });
  },
  updated(el) {
    const blocks = el.querySelectorAll("code");
    blocks.forEach((block) => {
      hljs.highlightElement(block);
    });
  },
});

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

app.use(pinia);
app.use(router);
app.use(naive);
app.use(ElementPlusX);
app.mount("#app");
