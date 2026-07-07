import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import '@mantine/nprogress/styles.css'

import './global.css'

import { Center, DirectionProvider, MantineProvider } from '@mantine/core'
import { QueryClientProvider } from '@tanstack/react-query'
import { NavigationProgress } from '@mantine/nprogress'
import { Notifications } from '@mantine/notifications'
import { ModalsProvider } from '@mantine/modals'
import { I18nextProvider } from 'react-i18next'
import { useMediaQuery } from '@mantine/hooks'
import { Suspense, useEffect } from 'react'

import { AuthProvider } from '@shared/hocs/auth-provider'
import { LoadingScreen } from '@shared/ui'
import { theme } from '@shared/constants'

import { Router } from './app/router/router'
import { queryClient } from './shared/api'
import i18n from './app/i18n/i18n'

export function App() {
    const mq = useMediaQuery('(min-width: 40em)')

    useEffect(() => {
        const root = document.getElementById('root')
        if (root && !root.querySelector('.safe-area-bottom')) {
            const bottomBar = document.createElement('div')
            bottomBar.className = 'safe-area-bottom'
            root.appendChild(bottomBar)
        }
    }, [])

    return (
        <I18nextProvider defaultNS="exodus" i18n={i18n}>
            <QueryClientProvider client={queryClient}>
                <AuthProvider>
                    <DirectionProvider>
                        <MantineProvider defaultColorScheme="dark" theme={theme}>
                            <ModalsProvider>
                                <Notifications position={mq ? 'top-right' : 'bottom-right'} />
                                <NavigationProgress />
                                <Suspense
                                    fallback={
                                        <Center h="100%">
                                            <LoadingScreen height="60vh" />
                                        </Center>
                                    }
                                >
                                    <Router />
                                </Suspense>
                            </ModalsProvider>
                        </MantineProvider>
                    </DirectionProvider>
                </AuthProvider>
            </QueryClientProvider>
        </I18nextProvider>
    )
}
