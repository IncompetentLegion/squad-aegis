export default defineNuxtRouteMiddleware(async (to) => {
  try {
    if (to.path === "/login") return;

    const redirectToLogin = () => {
      const redirectPath = to.fullPath;

      return navigateTo({
        path: "/login",
        query: { redirect: redirectPath },
      });
    };

    const authStore = useAuthStore();
    if (authStore.isLoggedIn) return;

    const authenticated = await authStore.fetch();
    if (!authenticated) return redirectToLogin();
  } catch (e) {
    return navigateTo({ path: "/login" });
  }
});
