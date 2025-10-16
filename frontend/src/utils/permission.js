/**
 * 权限工具函数
 * 用于检查用户权限级别
 * 权限级别：t0 → t1 → t2 → t3 → t4 → t5 → viewer → admin
 */

/**
 * 检查用户是否拥有指定权限级别
 * @param {Object} user - 用户对象
 * @param {string} requiredLevel - 需要的权限级别
 * @returns {boolean}
 */
export function hasPermission(user, requiredLevel) {
  if (!user || !user.groups) {
    return false
  }

  const groups = user.groups.toLowerCase()
  const required = requiredLevel.toLowerCase()
  return groups.includes(required)
}

/**
 * 检查用户是否具有审核权限（t4或admin）
 * @param {Object} user - 用户对象
 * @returns {boolean}
 */
export function isModerator(user) {
  return hasPermission(user, 't4') || hasPermission(user, 'admin')
}

/**
 * 检查用户是否具有管理员权限
 * @param {Object} user - 用户对象
 * @returns {boolean}
 */
export function isAdmin(user) {
  return hasPermission(user, 'admin')
}

/**
 * 获取用户权限级别显示名称
 * @param {Object} user - 用户对象
 * @returns {string}
 */
export function getPermissionLabel(user) {
  if (!user || !user.groups) {
    return '普通用户'
  }

  if (hasPermission(user, 'admin')) {
    return '管理员'
  }

  if (hasPermission(user, 't5')) {
    return 'T5'
  }

  if (hasPermission(user, 't4')) {
    return '审核员'
  }

  if (hasPermission(user, 'viewer')) {
    return '查看者'
  }

  // 提取最高的t级别
  const groups = user.groups.toLowerCase()
  for (let i = 5; i >= 0; i--) {
    if (groups.includes(`t${i}`)) {
      return `T${i}`
    }
  }

  return '普通用户'
}
