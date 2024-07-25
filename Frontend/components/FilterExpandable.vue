<template>
    <v-card class="mx-auto" elevation="4" outlined>
        <v-card-actions>
            <v-chip-group filter class="ml-3" v-model="name">
                <v-chip size="large" @click="emit('toggle-selection', props.name)" :value="props.name">{{ props.name
                    }}</v-chip>
            </v-chip-group>
            <v-spacer></v-spacer>
            <v-btn :icon="expand ? 'mdi-chevron-up' : 'mdi-chevron-down'" @click="expand = !expand"></v-btn>
        </v-card-actions>

        <v-expand-transition>
            <div v-show="expand">
                <v-divider></v-divider>
                <v-chip-group filter class="ml-2 pa-3" multiple v-model="highlight">
                    <v-chip v-for="(coordinate, index) in props.coordinates" class="mr-2"
                        :value="coordinate.bounding_id" @click="emit('toggle-highlight',coordinate.bounding_id)">
                        {{ props.name }} {{ index + 1}} ({{ (coordinate.confidence * 100).toFixed(1) }}%)
                    </v-chip>
                </v-chip-group>
            </div>
        </v-expand-transition>
    </v-card>
</template>

<script setup>
const expand = ref(false);
const props = defineProps(['name', 'coordinates'])
const emit = defineEmits(['toggle-selection','toggle-highlight']);
const name = ref(props.name);
const highlight = ref();
</script>

<style scoped></style>