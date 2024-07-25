<template>
    <div class="main my-4 ml-4 pa-8">
        <div class="font-weight-bold text-h6 mb-4">
            History
        </div>
        <div class="d-flex w-100 align-center justify-space-between">
            <div class="d-flex w-100">
                <div style="width: 40%;">
                    <v-text-field prepend-inner-icon="mdi-magnify" density="compact" label="Search image title"
                        variant="outlined" hide-details single-line v-model="searchName"
                        @click:prepend-inner="appendName" clearable></v-text-field>
                </div>
                <div style="width: 140px;" class="mx-4">
                    <v-select clearable label="Sort By" v-model="sortBy" :items="[
                        { text: 'Name', value: 'imageName' },
                        { text: 'Date', value: 'createdAt' },
                    ]" item-title="text" item-value="value" variant="outlined" density="compact" hide-details />
                </div>
                <div v-if="sortBy">
                    <v-btn icon="mdi-arrow-up" density="comfortable" variant="plain" class="border"
                        v-if="orderBy == 'asc'" @click="orderBy = 'desc'" />
                    <v-btn icon="mdi-arrow-down" density="comfortable" variant="plain" class="border"
                        v-if="orderBy == 'desc'" @click="orderBy = 'asc'" />
                </div>
            </div>
            <div>
                <v-btn-toggle v-model="display" divided density="comfortable" mandatory class="border">
                    <v-btn value="list"> <v-icon>mdi-format-list-bulleted</v-icon> </v-btn>
                    <v-btn value="grid"> <v-icon>mdi-view-grid-outline</v-icon></v-btn>
                </v-btn-toggle>
            </div>
        </div>
        <div class="mt-6 d-flex align-center">
            <v-btn icon="mdi-close" density="comfortable" variant="plain" @click="selectedFiles = []" class="mr-2"
                :disabled="selectedFiles.length <= 0" />
            <div class="mr-2">
                {{ selectedFiles.length }} item selected
            </div>
            <v-btn icon="mdi-trash-can-outline" density="comfortable" variant="plain" @click="deleteFiles"
                :disabled="selectedFiles.length <= 0" />
            <v-btn icon="mdi-file-export-outline" density="comfortable" variant="plain" @click="exportFile"
                :disabled="selectedFiles.length <= 0" />
        </div>
        <v-table v-if="display == 'list'" class="my-4">
            <thead>
                <tr>
                    <th class="text-left">
                        ID
                    </th>
                    <th class="text-left">
                        Image Title
                    </th>
                    <th class="text-left">
                        Status
                    </th>
                    <th class="text-left">
                        Date
                    </th>
                    <th class="text-left">
                        Image
                    </th>
                    <th></th>
                </tr>
            </thead>
            <tbody v-if="images">
                <!-- <ListImageCard v-for="file in fileStore.files" :id="file.id" :name="file.name" :preview="file.preview"
                    :status="file.status" :key="file.id" @toggle-selection="toggleSelection" :selectedId="selectedFiles"
                    @remove-selection="selectedFiles = []" /> -->
                <ListImageCard v-for="image in images.data" :id="image._id" :name="image.imageName"
                    :preview="image.imagePath" :status="image.status" :key="image._id" :date="image.createdAt"
                    @toggle-selection="toggleSelection" :selectedId="selectedFiles" @remove-selection="removeSelection"
                    @rename="fetchImages" />
            </tbody>
        </v-table>
        <div v-if="display == 'grid'">
            <v-row class="my-4" v-if="images">
                <!-- <v-col v-for="file in fileStore.files" :key="file.id" cols="12" sm="6" md="3">
                    <GridImageCard :img="file.preview" :name="file.name" :id="file.id"
                        @toggle-selection="toggleSelection" :selectedId="selectedFiles"
                        @remove-selection="selectedFiles = []" />
                </v-col> -->
                <v-col v-for="image in images.data" :key="image._id" cols="12" sm="6" md="3">
                    <GridImageCard :img="image.imagePath" :name="image.imageName" :id="image._id"
                        @toggle-selection="toggleSelection" :selectedId="selectedFiles"
                        @remove-selection="removeSelection" @rename="fetchImages" />
                </v-col>
            </v-row>
        </div>
    </div>
</template>

<script setup>
definePageMeta({
    middleware: 'auth'
})

const searchName = ref();
const sortBy = ref();
const orderBy = ref("desc");
const display = ref("list");
const selectedFiles = ref([]);
const imageName = ref(null)

const { token, data: user } = useAuth()
const images = ref(null)

const appendName = () => {
    imageName.value = searchName.value;
}

const queryParams = computed(() => {
    const params = new URLSearchParams()
    if (imageName.value) params.append('search', imageName.value)
    if (sortBy.value) params.append('sortBy', sortBy.value)
    if (orderBy.value) params.append('sortOrder', orderBy.value)
    return params.toString()
})

const fetchImages = async () => {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images?${queryParams.value}`, {
                method: 'GET',
                headers: {
                    'Authorization': token.value
                }
            })
            if (response.ok) {
                images.value = await response.json()
            } else {
                console.error('Failed to fetch image:', response.statusText)
            }
        } catch (error) {
            console.error('Error fetching image:', error)
        }
    }
}

watchEffect(() => {
    if (token.value) {
        console.log("fetch initiated");
        fetchImages();
    } else {
        console.log("fetch not intiated");
    }
})

function toggleSelection(id) {
    if (selectedFiles.value.includes(id)) {
        selectedFiles.value = selectedFiles.value.filter(fileId => fileId !== id);
    } else {
        selectedFiles.value.push(id);
    }
}

function removeSelection() {
    selectedFiles.value = [];
    fetchImages();
}

// function deleteFiles() {
//     fileStore.deleteManyFiles(selectedFiles.value);
//     selectedFiles.value = [];
// }

async function deleteFiles() {
    if (token.value) {
        const payload = {
            ids: selectedFiles.value,
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
                selectedFiles.value = [];
                fetchImages();
            } else {
                console.error('Failed to delete images:', response.statusText)
            }
        } catch (error) {
            console.error('Error deleting images:', error)
        }
    }
}


async function exportFile() {
    if (selectedFiles.value.length > 1) {
        exportManyImage();
    } else {
        const selectedImage = images.value.data.find(image => image._id === selectedFiles.value[0]);
        if (selectedImage) {
            exportImage(selectedImage.imageName);
        }
    }
}

async function exportImage(imageName) {
    if (token.value) {
        try {
            const response = await fetch(`http://localhost:8080/images/download/${selectedFiles.value[0]}`, {
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
            link.setAttribute('download', imageName);
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
            ids: selectedFiles.value,
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

watch([display], () => {
    selectedFiles.value = [];
})
</script>

<style scoped>
.main {
    background-color: white;
}
</style>