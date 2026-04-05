import type { Template } from './types'

export function getTemplateName(template: Template): string {
  return template.Name || template.name || 'Unnamed Template'
}

export function getTemplateSlug(template: Template): string {
  return template.Slug || template.slug || ''
}

export function getTemplateDescription(template: Template): string {
  return template.Description || template.description || 'Professional research computing environment'
}

export function getTemplateTags(template: Template): string[] {
  const tags: string[] = []

  if (template.category) {
    tags.push(template.category)
  }
  if (template.complexity) {
    tags.push(template.complexity)
  }
  if (template.package_manager) {
    tags.push(template.package_manager)
  }
  if (template.features && Array.isArray(template.features)) {
    tags.push(...template.features.slice(0, 2))
  }

  return tags
}
