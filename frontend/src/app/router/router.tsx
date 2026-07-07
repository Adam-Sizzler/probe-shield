import { Navigate, Outlet, Route, BrowserRouter, Routes, useNavigate } from 'react-router-dom'

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
