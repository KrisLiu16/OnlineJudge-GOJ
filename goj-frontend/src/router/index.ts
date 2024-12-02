import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import { useUserStore } from '@/stores/modules/user'
import NotFoundView from '@/views/NotFoundView.vue'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
  },
  {
    path: '/sign-in',
    name: 'sign-in',
    component: () => import('@/views/auth/SignInView.vue'),
  },
  {
    path: '/sign-up',
    name: 'sign-up',
    component: () => import('@/views/auth/SignUpView.vue'),
  },
  {
    path: '/problems',
    name: 'problems',
    component: () => import('@/views/problem/ProblemsView.vue'),
  },
  {
    path: '/problem/:id',
    name: 'problem-detail',
    component: () => import('@/views/problem/ProblemDetailView.vue'),
  },
  {
    path: '/contests',
    name: 'contests',
    component: () => import('@/views/contest/ContestsView.vue'),
  },
  {
    path: '/contest/:id',
    name: 'contest-detail',
    component: () => import('@/views/contest/ContestDetailView.vue'),
  },
  {
    path: '/discuss',
    name: 'discuss',
    component: () => import('@/views/discuss/DiscussView.vue'),
  },
  {
    path: '/discuss/new',
    name: 'discuss-new',
    component: () => import('@/views/discuss/DiscussEditView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/discuss/:id',
    name: 'discuss-detail',
    component: () => import('@/views/discuss/DiscussDetailView.vue'),
  },
  {
    path: '/discuss/edit/:id?',
    name: 'discuss-edit',
    component: () => import('@/views/discuss/DiscussEditView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/rank/:page?',
    name: 'rank',
    component: () => import('@/views/rank/RankView.vue'),
  },
  {
    path: '/submissions',
    name: 'submissions',
    component: () => import('@/views/submission/SubmissionView.vue'),
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: '/submission/:id',
    name: 'submission-detail',
    component: () => import('@/views/submission/SubmissionDetailView.vue'),
    meta: {
      requiresAuth: true,
    },
  },
  {
    path: '/groups',
    name: 'groups',
    component: () => import('@/views/group/GroupsView.vue'),
  },
  {
    path: '/profile/:username?',
    name: 'profile',
    component: () => import('@/views/user/UserProfileView.vue'),
    props: true,
    meta: { requiresAuth: false },
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/user/components/UserSettings.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'admin',
    component: () => import('@/views/admin/index.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      hideFooter: true,
    },
    children: [
      {
        path: '',
        component: () => import('@/views/admin/dashboard/Index.vue'),
      },
      {
        path: 'problem/add',
        component: () => import('@/views/admin/problem/AddProblem.vue'),
      },
      {
        path: 'problem/manage',
        component: () => import('@/views/admin/problem/ManageProblems.vue'),
      },
      {
        path: 'problem/edit/:id',
        component: () => import('@/views/admin/problem/EditProblem.vue'),
      },
      {
        path: 'problem/data/:id',
        component: () => import('@/views/admin/problem/ManageData.vue'),
      },
      {
        path: 'contest/add',
        component: () => import('@/views/admin/contest/AddContest.vue'),
      },
      {
        path: 'contest/manage',
        component: () => import('@/views/admin/contest/ManageContests.vue'),
      },
      {
        path: 'contest/edit/:id',
        component: () => import('@/views/admin/contest/EditContest.vue'),
      },
      {
        path: 'users',
        component: () => import('@/views/admin/user/ManageUsers.vue'),
      },
      {
        path: 'roles',
        component: () => import('@/views/admin/user/ManageRoles.vue'),
      },
      {
        path: 'problem/import-export',
        component: () => import('@/views/admin/problem/ImportExport.vue'),
      },
      {
        path: 'website/basic',
        component: () => import('@/views/admin/website/basic.vue'),
        meta: { requiresAdmin: true }
      },
    ],
  },
  {
    path: '/contest/:contestId/problem/:problemIndex',
    name: 'contest-problem',
    component: () => import('@/views/contest/ContestProblemView.vue'),
    meta: {
      hideFooter: true,
    },
    props: (route) => ({
      contestId: route.params.contestId,
      problemIndex: route.params.problemIndex,
    }),
  },
  {
    path: '/contest/:id/rank',
    name: 'contest-rank',
    component: () => import('@/views/contest/ContestRankView.vue'),
  },
  {
    path: '/contest/:id/participants',
    name: 'contest-participants',
    component: () => import('@/views/contest/ContestParticipantsView.vue'),
  },
  {
    path: '/about',
    name: 'about',
    component: () => import('@/views/AboutView.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFoundView,
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAdmin && userStore.userInfo?.role !== 'admin') {
    next('/')
    return
  }

  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    next({ name: 'sign-in', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
