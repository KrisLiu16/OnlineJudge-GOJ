declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $confirm: (options?: {
      title?: string
      message?: string
      type?: string
      confirmText?: string
      cancelText?: string
    }) => Promise<boolean>
  }
}
