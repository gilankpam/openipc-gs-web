import { createTheme, type MantineColorsTuple } from '@mantine/core';

const brand: MantineColorsTuple = [
    '#e5f4ff',
    '#cde2ff',
    '#9bc2ff',
    '#64a0ff',
    '#3984fe',
    '#1d72fe',
    '#0969ff',
    '#0058e4',
    '#004ecc',
    '#0043b5'
];

export const theme = createTheme({
    primaryColor: 'brand',
    colors: {
        brand,
    },
    components: {
        Paper: {
            defaultProps: {
                bg: 'dark.7',
            }
        }
    }
});
