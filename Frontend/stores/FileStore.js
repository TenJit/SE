import { defineStore } from 'pinia'

export const useFileStore = defineStore('fileStore',{
    state: () => ({
        files:[

        ],
        id:1
    }),
    actions: {
        addFile(file){
            this.files.push(file)
        },
        deleteFile(id){
            this.files = this.files.filter(f => {
                return f.id !== id
            })
        },
        deleteManyFiles(ids){
            this.files = this.files.filter(f => !ids.includes(f.id));
        },
        changeFileName(id,name){
            const file = this.files.find(f => f.id === id)
            file.name = name;
        }
    },
    getters: {
        getFileById: (state) => (id) => {
            return state.files.find((file) => file.id === id);
        },
    },
})