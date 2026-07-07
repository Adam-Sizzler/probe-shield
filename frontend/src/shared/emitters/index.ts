type Listener = () => void

function createEmitter() {
    const listeners = new Set<Listener>()

    return {
        emit: () => listeners.forEach((listener) => listener()),
        subscribe: (listener: Listener) => {
            listeners.add(listener)
            return () => listeners.delete(listener)
        }
    }
}

export const logoutEvents = createEmitter()
