import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
    {
        path: '/',
        component: () => import('../layout/MainLayout.vue'),
        children: [
            {
                path: '',
                name: 'Home',
                component: () => import('../views/Home.vue')
            },
            {
                path: 'projects',
                name: 'Projects',
                component: () => import('../views/Projects.vue')
            },
            {
                path: 'models',
                name: 'Models',
                component: () => import('../views/Models.vue')
            },
            {
                path: 'scripts',
                name: 'Scripts',
                component: () => import('../views/Scripts.vue')
            },
            {
                path: 'agents',
                name: 'Agents',
                component: () => import('../views/Agents.vue')
            },
            {
                path: 'tools',
                name: 'Tools',
                component: () => import('../views/Tools.vue')
            },
            {
                path: 'chat/:sessionId?',
                name: 'Chat',
                component: () => import('../views/Chat.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHashHistory(),
    routes
})

export default router
