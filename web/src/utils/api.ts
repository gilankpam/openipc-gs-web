interface RequestInitWithTimeout extends RequestInit {
    timeout?: number;
}

export const fetchWithTimeout = async (resource: RequestInfo, options: RequestInitWithTimeout = {}) => {
    const { timeout = 2000, ...fetchOptions } = options;

    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);

    const response = await fetch(resource, {
        ...fetchOptions,
        signal: controller.signal
    });

    clearTimeout(id);
    return response;
};
