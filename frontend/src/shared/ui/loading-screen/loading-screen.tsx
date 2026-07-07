import { Loader, Stack } from '@mantine/core'

export function LoadingScreen({ height = '100vh' }: { height?: string | number }) {
    return (
        <Stack align="center" h={height} justify="center">
            <Loader color="cyan" />
        </Stack>
    )
}
