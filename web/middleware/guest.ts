export default defineNuxtRouteMiddleware(async () => {
  const authStore = useAuthStore();

  if (authStore.isLoggedIn || await authStore.fetch()) {
    return navigateTo("/dashboard");
  }
});
