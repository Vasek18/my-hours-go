import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/WeekCalendarView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/activity-types',
      name: 'activity-types',
      component: () => import('@/views/ActivityTypesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/activity-types/new',
      name: 'activity-type-new',
      component: () => import('@/views/ActivityTypeFormView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/activity-types/:id/edit',
      name: 'activity-type-edit',
      component: () => import('@/views/ActivityTypeFormView.vue'),
      meta: { requiresAuth: true },
    },
    {
      // Public: reachable from the confirmation link whether or not signed in.
      path: '/confirm-email',
      name: 'confirm-email',
      component: () => import('@/views/ConfirmEmailView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
    },
  ],
})

// Single source of truth for redirects based on session state.
router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // Resolve the session once, before the first guarded navigation.
  if (!auth.initialized) {
    await auth.fetchMe()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresGuest && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }

  return true
})

export default router
