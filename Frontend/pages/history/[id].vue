<template>
    <div class="main my-4 ml-4 pa-8">
        <div v-if="image">
            <div class="font-weight-bold text-h6 mb-3 d-flex justify-space-between">
                <div>
                    {{ image.data.imageName }}
                    <v-btn icon="mdi-pencil-outline" density="comfortable" variant="plain"
                        @click="editNamePopUp = true"></v-btn>
                </div>
                <div class="d-flex">
                    <v-btn icon="mdi-trash-can-outline" density="comfortable" variant="plain" @click="deleteImage" />
                    <v-btn icon="mdi-file-export-outline" density="comfortable" variant="plain" @click="exportImage" />
                </div>
            </div>
            <v-row class="my-2">
                <v-col cols="12" sm="12" md="7" class="image-container">
                    <img ref="imageElement" :src="'http://localhost:8080/' + image.data.imagePath" @load="onImageLoad"
                        class="main-image" />
                    <canvas ref="canvas" class="overlay-canvas"></canvas>
                </v-col>
                <v-col cols="12" sm="12" md="5">
                    <FilterExpandable class="mb-4" v-for="result in image.data.result" :name="result.name"
                        :coordinates="result.coordinates" @toggle-selection="toggleSelection"
                        @toggle-highlight="toggleHighlight" v-if="image.data.result" />
                    <div v-else-if="image.data.status == 'success'" class="text-h6 font-weight-black">No object detected
                    </div>
                    <div v-if="image.data.status == 'pending'" class="text-h6 font-weight-black">The image is still {{
                        image.data.status }}</div>
                    <div v-if="image.data.status == 'fail'" class="text-h6 font-weight-black">Fail to run detection
                    </div>
                </v-col>
            </v-row>
        </div>
    </div>
    <v-dialog overlay overlay-opacity="0.8" scrim="black" v-model="editNamePopUp">
        <v-fade-transition hide-on-leave>
            <v-card class="custom-card" elevation="16" width="40.5vw">
                <v-card-title class="d-flex justify-space-between align-center">
                    <div class="text-h5 text-medium-emphasis ps-2">
                        Change File Name
                    </div>
                    <v-btn icon="mdi-close" variant="text" @click="editNamePopUp = false"></v-btn>
                </v-card-title>
                <v-divider></v-divider>
                <v-text-field v-model="newName" :label="image.data.imageName" message="name" clearable hide-details
                    single-line>
                </v-text-field>
                <v-btn @click="saveNewName(newName)">
                    Save
                </v-btn>
            </v-card>
        </v-fade-transition>
    </v-dialog>
</template>


<script setup>

const { id } = useRoute().params;
const fileStore = useFileStore();
const file = fileStore.getFileById(Number(id));
const { token } = useAuth();

const editNamePopUp = ref(false);
const newName = ref(null);
const image = ref(null);
const selectedNames = ref([]);
const highlighted = ref([]);

function toggleSelection(name) {
    if (selectedNames.value.includes(name)) {
        selectedNames.value = selectedNames.value.filter(objName => objName !== name);
    } else {
        selectedNames.value.push(name);
    }
    drawBoundingBoxes();
}

function toggleHighlight(id) {
    if (highlighted.value.includes(id)) {
        highlighted.value = highlighted.value.filter(objId => objId !== id);
    } else {
        highlighted.value.push(id);
    }
    drawBoundingBoxes();
}

const fetchImage = async () => {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/${id}`, {
                method: 'GET',
                headers: {
                    'Authorization': token.value,
                },
            });
            if (response.ok) {
                image.value = await response.json();
                selectedNames.value = image.value.data.result.map(r => r.name);
                console.log('Fetched image:', image.value);
                await nextTick();
                drawBoundingBoxes();
            } else {
                console.error('Failed to fetch image:', response.statusText);
            }
        } catch (error) {
            console.error('Error fetching image:', error);
        }
    }
};

watchEffect(() => {
    if (token.value) {
        console.log('fetch initiated');
        fetchImage();
    } else {
        console.log('fetch not initiated');
    }
});

const onImageLoad = () => {
    drawBoundingBoxes();
};

const drawBoundingBoxes = () => {
    const imgElement = document.querySelector('.main-image');
    const canvas = document.querySelector('.overlay-canvas');
    const ctx = canvas.getContext('2d');

    const img = new Image();
    img.src = imgElement.src;
    img.onload = () => {
        const imgWidth = imgElement.clientWidth;
        const imgHeight = imgElement.clientHeight;

        console.log('Image dimensions:', imgWidth, imgHeight);

        canvas.width = imgWidth;
        canvas.height = imgHeight;

        const scaleX = imgWidth / img.width;
        const scaleY = imgHeight / img.height;

        console.log('Scale factors:', scaleX, scaleY);

        ctx.clearRect(0, 0, canvas.width, canvas.height);

        image.value.data.result.forEach((object) => {
            if (selectedNames.value.includes(object.name)) {
                object.coordinates.forEach((coord) => {
                    const x = coord.x_min * scaleX;
                    const y = coord.y_min * scaleY;
                    const width = (coord.x_max - coord.x_min) * scaleX;
                    const height = (coord.y_max - coord.y_min) * scaleY;

                    console.log('Drawing box:', x, y, width, height);

                    if (highlighted.value.includes(coord.bounding_id)) {
                        ctx.strokeStyle = 'red';
                        ctx.lineWidth = 4;
                        ctx.font = '16px Arial';
                        ctx.fillStyle = 'red';
                    } else {
                        ctx.strokeStyle = 'black';
                        ctx.lineWidth = 2;
                        ctx.font = '16px Arial';
                        ctx.fillStyle = 'black';
                    }

                    ctx.beginPath();
                    ctx.rect(x, y, width, height);
                    ctx.stroke();
                    ctx.closePath();
                    ctx.fillText(
                        `${object.name} (${(coord.confidence * 100).toFixed(1)}%)`,
                        x,
                        y - 5
                    );
                });
            }
        });
    };
};

async function saveNewName(name) {
    if (token.value) {
        try {
            const formData = new FormData();
            formData.append('name', name);
            const response = await fetch(`http://localhost:8080/images/${id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': token.value,
                },
                body: formData,
            });

            const data = await response.json();

            if (response.ok) {
                editNamePopUp.value = false;
                newName.value = '';
                fetchImage();
            } else {
                console.error('Error renaming image:', data.message || data.error);
            }
        } catch (error) {
            console.error('Error renaming image:', error);
        }
    } else {
        navigateTo('/signin');
    }
}

async function deleteImage() {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/${id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': token.value,
                },
            });
            if (response.ok) {
                navigateTo('/history');
            } else {
                console.error('Failed to delete image:', response.statusText);
            }
        } catch (error) {
            console.error('Error deleting images:', error);
        }
    }
}

async function exportImage() {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/download/${id}`, {
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
            link.setAttribute('download', image.value.data.imageName);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            alert(`Error downloading image: ${error.message}`);
        }
    }
}
</script>



<style scoped>
.main {
    background-color: white;
}

.custom-card {
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    position: fixed;
    z-index: 999;
}

.image-container {
    position: relative;
    width: 100%;
}

.main-image {
    width: 100%;
    height: auto;
    display: block;
    object-fit: cover;
}

.overlay-canvas {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
}
</style>