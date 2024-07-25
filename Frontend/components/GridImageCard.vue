<template>
    <v-card class="mx-auto" :color="cardColor" rounded hover @click="emitToggleSelection" :ripple="false"
        @dblclick="navigateTo(`/history/${props.id}`)">
        <div class="d-flex align-center py-1 justify-space-between">
            <div style="width: 85%;">
                <v-card-title class="text-subtitle-1 text-truncate">{{ props.name }}</v-card-title>
            </div>
            <v-menu location="bottom">
                <template v-slot:activator="{ props }">
                    <v-btn icon="mdi-dots-vertical" v-bind="props" density="compact" variant="plain"
                        class="mr-2"></v-btn>
                </template>
                <v-list>
                    <v-list-item>
                        <v-btn @click="editNamePopUp = true" variant="text" block
                            prepend-icon="mdi-pencil-outline">Rename</v-btn>
                    </v-list-item>
                    <v-list-item>
                        <v-btn @click="deleteFile()" variant="text" block
                            prepend-icon="mdi-trash-can-outline">Delete</v-btn>
                    </v-list-item>
                    <v-list-item>
                        <v-btn variant="text" block prepend-icon="mdi-file-export-outline"
                            @click="exportFile">Export</v-btn>
                    </v-list-item>
                </v-list>
            </v-menu>
        </div>
        <div class="px-2 pb-2">
            <v-img height="225" :src="imagePath" cover rounded lazy></v-img>
        </div>

    </v-card>
    <v-dialog overlay overlay-opacity="0.8" scrim="black" v-model="editNamePopUp">
        <v-fade-transition hide-on-leave><v-card class="custom-card" elevation="16" width="40.5vw">
                <v-card-title class="d-flex justify-space-between align-center">
                    <div class="text-h5 text-medium-emphasis ps-2">
                        Change File Name
                    </div>
                    <v-btn icon="mdi-close" variant="text" @click="editNamePopUp = false"></v-btn>
                </v-card-title>
                <v-divider></v-divider>
                <v-text-field v-model="newName" :label="props.name" message="name" clearable hide-details single-line>

                </v-text-field>
                <v-btn @click="saveNewName(props.id, newName)">
                    Save
                </v-btn>
            </v-card>
        </v-fade-transition>
    </v-dialog>
</template>

<script setup>
const props = defineProps(['img', 'name', 'id', 'status', 'selectedId'])
const emit = defineEmits(['toggle-selection', 'remove-selection', 'rename']);
const fileStore = useFileStore();
const editNamePopUp = ref(false);
const newName = ref(null);

const { token, data: user } = useAuth()

const imagePath = "http://localhost:8080/" + props.img;

const cardColor = computed(() => props.selectedId.includes(props.id) ? '#B1DEFF' : '#F1EFEF');

function emitToggleSelection() {
    emit('toggle-selection', props.id);
}

// function saveNewName(id,name){
//     fileStore.changeFileName(id,name);
//     editNamePopUp.value = false;
//     newName.value = "";
// }

async function saveNewName(id, name) {
    if (token.value) {
        try {
            const formData = new FormData();
            formData.append('name', name);
            const response = await fetch(`http://localhost:8080/images/${id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': token.value
                },
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                editNamePopUp.value = false;
                newName.value = "";
                emit('rename');
            } else {
                console.error('Error renaming image:', data.message || data.error);
            }
        } catch (error) {
            console.error('Error renaming image:', error);
        }
    } else {
        navigateTo('/signin')
    }
}

// function deleteImage(){
//     if(props.selectedId.length>1){
//         fileStore.deleteManyFiles(props.selectedId);
//     }else{
//         fileStore.deleteFile(props.id);
//     }
//     emit('remove-selection');
// }

async function deleteFile() {
    if (props.selectedId.length > 1) {
        deleteManyImage();
    } else {
        deleteImage();
    }
}

async function deleteManyImage() {
    if (token.value) {
        const payload = {
            ids: props.selectedId,
        };
        try {
            const response = await fetch('http://localhost:8080/images', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': token.value
                },
                body: JSON.stringify(payload)
            })
            if (response.ok) {
                emit('remove-selection');
            } else {
                console.error('Failed to delete images:', response.statusText)
            }
        } catch (error) {
            console.error('Error deleting images:', error)
        }
    }
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
                emit('remove-selection');
            } else {
                console.error('Failed to delete image:', response.statusText)
            }
        } catch (error) {
            console.error('Error deleting images:', error)
        }
    }
}


async function exportFile() {
    if (props.selectedId.length > 1) {
        exportManyImage();
    } else {
        exportImage();
    }
    emit('remove-selection');
}

async function exportImage() {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/download/${props.id}`, {
                method: 'GET',
                headers: {
                    'Authorization': token.value,
                },
            });

            if (!response.ok) {
                throw new Error('Failed to download image');
            }

            const blob = await response.blob();
            const blobUrl = window.URL.createObjectURL(blob);

            const link = document.createElement('a');
            link.href = blobUrl;
            link.setAttribute('download', `${props.name}`);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            alert(`Error downloading image: ${error.message}`);
        }
    }
}

async function exportManyImage() {
    if (token.value) {
        const payload = {
            ids: props.selectedId,
        };
        try {
            const response = await fetch('http://localhost:8080/images/downloadManyImages', {
                method: 'POST',
                headers: {
                    'Authorization': token.value,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                throw new Error('Failed to download images');
            }

            const blob = await response.blob();
            const blobUrl = window.URL.createObjectURL(blob);

            const currentDate = new Date();
            const year = currentDate.getFullYear();
            const month = String(currentDate.getMonth() + 1).padStart(2, '0');
            const day = String(currentDate.getDate()).padStart(2, '0');
            const hours = String(currentDate.getHours()).padStart(2, '0');
            const minutes = String(currentDate.getMinutes()).padStart(2, '0');
            const seconds = String(currentDate.getSeconds()).padStart(2, '0');
            const fileName = `${day}_${month}_${year}-${hours}_${minutes}_${seconds}.zip`;

            const link = document.createElement('a');
            link.href = blobUrl;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            alert(`Error downloading images: ${error.message}`);
        }
    }
};

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