<template>
    <v-layout class="rounded rounded-md">
      <v-navigation-drawer expand-on-hover rail :width="220">
        <v-list>
          <v-list-item
            class="py-4 font-weight-black"
            prepend-avatar="/logo.png"
            title="Image Detection"
          ></v-list-item>
        </v-list>
  
        <v-divider></v-divider>
  
        <v-list density="compact" nav>
          <v-list-item
            :to="{ name: 'index' }"
            prepend-icon="mdi-upload"
            title="Upload Image"
          ></v-list-item>
          <v-list-item
            :to="{ name: 'history' }"
            prepend-icon="mdi-history"
            title="History"
          ></v-list-item>
          <v-list-item
            :to="{ name: 'profile' }"
            prepend-icon="mdi-account"
            title="Profile"
          ></v-list-item>
        </v-list>
      </v-navigation-drawer>
      <v-main class="main">
        <slot/>
      </v-main>
    </v-layout>
  </template>
  
  <script setup>
  const route = useRoute();
  </script>
  
  <style scoped>
    .main{
        background-color: #f2f2f2;
        min-height: 100dvh;
        min-width: 100dvw;
    }
  </style>
  