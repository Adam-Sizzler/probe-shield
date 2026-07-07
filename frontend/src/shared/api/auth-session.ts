export async function logout(): Promise<void> {
    try {
        await fetch('/api/auth/logout', {
            method: 'POST',
            credentials: 'include'
        })
    } catch {
        // best-effort logout
    }
}
