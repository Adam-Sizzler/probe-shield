import { createContext, ReactNode, useEffect, useMemo, useState } from 'react'

import { removeToken, useToken } from '@entities/auth'
import { logoutEvents } from '@shared/emitters'

interface AuthContextValues {
    isAuthenticated: boolean
    isInitialized: boolean
    setIsAuthenticated: (isAuthenticated: boolean) => void
}

export const AuthContext = createContext<AuthContextValues | null>(null)

interface AuthProviderProps {
    children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [isAuthenticated, setIsAuthenticated] = useState(false)
    const [isInitialized, setIsInitialized] = useState(false)
    const token = useToken()

    const logoutUser = () => {
        setIsAuthenticated(false)
        removeToken()
    }

    useEffect(() => {
        const unsubscribe = logoutEvents.subscribe(() => {
            logoutUser()
        })

        return () => {
            unsubscribe()
        }
    }, [])

    useEffect(() => {
        setIsAuthenticated(Boolean(token))
        setIsInitialized(true)
    }, [])

    const value = useMemo(
        () => ({ isAuthenticated, isInitialized, setIsAuthenticated }),
        [isAuthenticated, isInitialized]
    )

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
