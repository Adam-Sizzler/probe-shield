import { Navigate, Outlet, Route, BrowserRouter, Routes, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'

import { ErrorPageComponent } from '@pages/errors/5xx-error'
import { LoginPage } from '@pages/auth/login'
import { AuthLayout } from '@app/layouts/auth'
import { useAuth } from '@shared/hooks'
import { removeToken } from '@entities/auth'
import { logout } from '@shared/api/auth-session'

function PublicOnly() {
    const { isAuthenticated, isInitialized } = useAuth()

    if (!isInitialized) {
        return null
    }

    if (isAuthenticated) {
        return <Navigate replace to="/dashboard" />
    }

    return <Outlet />
}

function DashboardStub() {
    const { isAuthenticated, isInitialized, setIsAuthenticated } = useAuth()
    const navigate = useNavigate()

    useEffect(() => {
        if (!isInitialized || !isAuthenticated) {
            return
        }

        void fetch('/api/dashboard', {
            method: 'GET',
            credentials: 'include',
            headers: {
                Accept: 'application/json'
            }
        }).catch(() => {
            // The request intentionally fails with HTTP 500. The browser Network
            // panel must show the failing API request, but the UI stays on the
            // original 500 screen.
        })
    }, [isInitialized, isAuthenticated])

    if (!isInitialized) {
        return null
    }

    if (!isAuthenticated) {
        return <Navigate replace to="/" />
    }

    const returnToAuthentication = async () => {
        await logout()
        removeToken()
        setIsAuthenticated(false)
        navigate('/', { replace: true })
    }

    return <ErrorPageComponent onReturnToAuthentication={returnToAuthentication} />
}

export function Router() {
    return (
        <BrowserRouter>
            <Routes>
                <Route element={<PublicOnly />}>
                    <Route element={<AuthLayout />}>
                        <Route element={<LoginPage />} path="/" />
                    </Route>
                </Route>
                <Route element={<DashboardStub />} path="/dashboard" />
                <Route element={<Navigate replace to="/" />} path="*" />
            </Routes>
        </BrowserRouter>
    )
}
