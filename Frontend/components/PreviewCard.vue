<template>
    <div class="mb-6 d-flex">
        <v-img width="37.5%" height="150" :src="props.imgurl" lazy cover />
        <div style="width: 65%;" class="text-h6 pl-8 pr-4 d-flex flex-column justify-space-between">
            <div class="text-truncate d-flex align-center">
                <div style="max-width: 85%;" class="text-truncate">
                    Name: {{ props.name }}&nbsp;
                </div>
                <v-icon icon="mdi-pencil-outline" size="x-small" v-if="props.status == 'success'"
                    @click="editNamePopUp = true"></v-icon>
            </div>
            <div>
                Status: {{ props.status }} 
                <v-icon color="#F24E1E" icon="mdi-alert-circle" size="small" v-if="props.status=='fail'"/>
                <v-icon color="#699BF7" icon="mdi-upload-circle" size="small" v-if="props.status=='pending'"/>
                <v-icon color="#0FA958" icon="mdi-check-circle" size="small" v-if="props.status=='success'"/>
            </div>
            <div>
                <v-btn variant="flat" color="red" prepend-icon="mdi-close-circle-outline"
                    v-if="props.status == 'pending'" @click="deleteImage">
                    Cancel
                </v-btn>
                <v-btn variant="flat" color="red" prepend-icon="mdi-close-circle-outline" v-if="props.status == 'fail'"
                    @click="deleteImage">
                    Remove
                </v-btn>
                <NuxtLink :to="`/history/${props.id}`">
                    <v-btn variant="flat" color="green" v-if="props.status == 'success'">
                    View Result
                </v-btn>
                
                </NuxtLink>
                
            </div>
        </div>
        <v-dialog overlay overlay-opacity="0.8" scrim="black" v-model="editNamePopUp">
            <v-fade-transition hide-on-leave><v-card class="custom-card" elevation="16" width="40.5vw">
                    <v-card-title class="d-flex justify-space-between align-center">
                        <div class="text-h5 text-medium-emphasis ps-2">
                            Change File Name
                        </div>
                        <v-btn icon="mdi-close" variant="text" @click="editNamePopUp = false"></v-btn>
                    </v-card-title>
                    <v-divider></v-divider>
                    <v-text-field v-model="newName" :label="props.name" message="name" clearable hide-details
                        single-line>

                    </v-text-field>
                    <v-btn @click="saveNewName(props.id, newName)">
                        Save
                    </v-btn>
                </v-card>
            </v-fade-transition>
        </v-dialog>

    </div>
</template>

<script setup>
const props = defineProps(['name', 'status', 'imgurl', 'id', 'status']);
const editNamePopUp = ref(false);
const newName = ref(null);
const { token, data: user } = useAuth();
const emit = defineEmits(['delete']);

function saveNewName(id, name) {
    fileStore.changeFileName(id, name);
    editNamePopUp.value = false;
    newName.value = "";
}

async function deleteImage() {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/${props.id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': token.value
                }
            })
            if (response.ok) {
                emit('delete');
            } else {
                console.error('Failed to delete image:', response.statusText)
            }
        } catch (error) {
            console.error('Error deleting images:', error)
        }
    }
}


</script>

<style scoped>
.custom-card {
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    position: fixed;
    z-index: 999;
}
</style>
