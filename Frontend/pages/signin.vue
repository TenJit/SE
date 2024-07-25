<template>
    <div>
        <v-form validate-on="submit lazy" @submit.prevent="login" style="height: 100vh;" class="d-flex align-center">
            <v-card class="mx-auto pa-12 pb-8 my-auto" elevation="8" min-width="35%" rounded="lg">
                <div class="text-subtitle-1 text-medium-emphasis">Account</div>

                <v-text-field density="compact" placeholder="Email address" prepend-inner-icon="mdi-email-outline"
                    variant="outlined" v-model="formData.email" name="email"></v-text-field>

                <div class="text-subtitle-1 text-medium-emphasis d-flex align-center justify-space-between">
                    Password
                </div>

                <v-text-field :append-inner-icon="passswordVisible ? 'mdi-eye-off' : 'mdi-eye'"
                    :type="passswordVisible ? 'text' : 'password'" density="compact" placeholder="Enter your password"
                    prepend-inner-icon="mdi-lock-outline" variant="outlined" v-model="formData.password"
                    @click:append-inner="passswordVisible = !passswordVisible"></v-text-field>

                <v-btn class="mb-8" color="blue" size="large" variant="tonal" block type="submit">
                    Log In
                </v-btn>
                <div class="d-flex justify-center">
                    Not a member?&nbsp;
                    <NuxtLink to="/signup" class="no-decoration">
                        Sign Up Now!
                    </NuxtLink>
                </div>
            </v-card>
        </v-form>

    </div>
</template>

<script setup>

const { signIn } = useAuth() // uses the default signIn function provided by nuxt-auth
const formData = reactive({
    email: '',
    password: '',
})

const passswordVisible = ref(false);

const login = async (e) => {
    try {
        e.preventDefault()
        let res = await signIn(
            { ...formData },
            { callbackUrl: '/' } // Where the user will be redirected after a successiful login
        )

        console.log("res", res);

    } catch (error) {
        console.log("error", error);
        alert("wrong email/password");
    }
}

definePageMeta({
    title: 'Signin',
    layout: 'empty',
    public: true,
    auth: {
        unauthenticatedOnly: true,
        navigateAuthenticatedTo: '/',
    },
})


</script>

<style scoped></style>