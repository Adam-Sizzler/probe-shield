import { notifications } from '@mantine/notifications'

export function errorHandler(error: unknown, title: string) {
    const message = error instanceof Error ? error.message : 'Request failed with unknown error.'

    notifications.show({
        title,
        message,
        color: 'red'
    })
}
