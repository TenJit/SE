// https://nuxt.com/docs/api/configuration/nuxt-config

import vuetify, { transformAssetUrls } from "vite-plugin-vuetify";

export default defineNuxtConfig({
  devtools: { enabled: true },

  css: ["vuetify/styles/main.sass", "@mdi/font/css/materialdesignicons.css"],
  build: {
    transpile: ["vuetify"],
  },
  modules: [
    "@pinia/nuxt",
    (_options, nuxt) => {
      nuxt.hooks.hook("vite:extendConfig", (config) => {
        // @ts-expect-error
        config.plugins.push(vuetify({ autoImport: true }));
      });
    },
    "@sidebase/nuxt-auth",
    //...
  ],
  auth: {
    baseURL: "http://localhost:8080/",
    provider: {
      type: "local",
      endpoints: {
        signIn: { path: "login", method: "post" },
        signOut: { path: "logout", method: "get" },
        signUp: { path: "register", method: "post" },
        getSession: { path: "getme", method: "get" },
      },
      pages: {
        login: "/signin",
      },
      token: {
        signInResponseTokenPointer: "/token",
      },
    },
  },
  vite: {
    vue: {
      template: {
        transformAssetUrls,
      },
    },
  },
});
