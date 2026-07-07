export function createUrl(endpoint: string, query?: Record<string, unknown>, route?: Record<string, string>) {
    let url = endpoint

    if (route) {
        Object.entries(route).forEach(([key, value]) => {
            url = url.replace(`:${key}`, value)
        })
    }

    if (query && Object.keys(query).length > 0) {
        const params = new URLSearchParams()
        Object.entries(query).forEach(([key, value]) => {
            if (value !== undefined && value !== null) {
                params.set(key, String(value))
            }
        })
        const search = params.toString()
        if (search) {
            url += `?${search}`
        }
    }

    return url
}
