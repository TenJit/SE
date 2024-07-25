<template>
    <div class="main my-4 ml-4 pa-8">
        <div class="font-weight-bold text-h6 mb-4">
            Upload your image
        </div>
        <div class="d-flex">
            <div class="dropfile d-flex flex-column align-center justify-center flex-grow-1" @drop.prevent="handleDrop"
                @dragover.prevent>
                <div class="font-weight-bold">
                    Drag and Drop file here
                </div>
                <div style="color: #A1a1a1;">
                    OR
                </div>

                <v-btn prepend-icon="mdi-plus-circle-outline" variant="flat" size="small" color="#FF8B36"
                    @click="selectFiles">
                    Add files
                </v-btn>
                <input type="file" ref="fileInput" @change="handleFileChange" multiple style="display: none;"
                    accept="image/*">
            </div>
            <div class="files pl-8 flex-grow-1" v-if="images && images.counts > 0">
                <PreviewCard v-for="image in images.data" :key="image._id" :imgurl="'http://localhost:8080/' + image.imagePath" :name="image.imageName"
                    :id="image._id" :status="image.status" @delete="fetchImages()"/>
            </div>
            
        </div>
    </div>
</template>

<script setup>

const { token, data: user } = useAuth();
const images = ref(null);

const fileInput = ref(null);

const selectFiles = () => {
    if (!user.value) {
        navigateTo('/signin');
    } else {
        fileInput.value.click();
    }
};

const handleFileChange = (event) => {
    const newFiles = Array.from(event.target.files);
    uploadFiles(newFiles);
};

const handleDrop = (event) => {
    const newFiles = Array.from(event.dataTransfer.files);
    uploadFiles(newFiles);
};

const uploadFiles = async (newFiles) => {
    for (const file of newFiles) {
        if (file.type.startsWith('image/')) {
            const formData = new FormData();
            formData.append('image', file);
            formData.append('name', file.name);

            try {
                const response = await fetch('http://localhost:8080/images', {
                    method: 'POST',
                    headers: { 'Authorization': token.value },
                    body: formData,
                });

                if (!response.ok) {
                    throw new Error('Failed to upload image');
                }else{
                    
                }
            } catch (error) {
                alert(`Error uploading ${file.name}: ${error.message}`);
            }
            fetchImages();
        } else {
            alert(`${file.name} is not an image file and will not be added.`);
        }
    }
};

const fetchImages = async () => {
    if (token.value) {
        try {
            const response = await fetch('http://localhost:8080/images?sortBy=createdAt&sortOrder=desc&recent=true', {
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
</script>

<style scoped>
.main {
    background-color: white;
}

.dropfile {
    background-color: #F2F2F2;
    height: 400px;
    width: 50%;
}

.files {
    width: 50%;
}
</style>
