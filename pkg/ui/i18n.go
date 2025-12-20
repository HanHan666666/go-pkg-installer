// Package ui provides simple i18n helpers for built-in UI strings.
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

var uiTranslations = map[string]map[string]string{
	"en": {
		"button.back":             "Go Back",
		"button.cancel":           "Cancel",
		"button.continue":         "Continue",
		"button.install":          "Install",
		"button.finish":           "Finish",
		"button.exit":             "Exit",
		"button.close":            "Close",
		"button.browse":           "Browse...",
		"dialog.cancel.title":     "Cancel Installation",
		"dialog.cancel.msg":       "Are you sure you want to cancel the installation?",
		"dialog.select.dir":       "Select Directory",
		"dialog.select.file":      "Select File",
		"dialog.validation.title": "Validation Error",
		"dialog.error.title":      "Error",
		"title.welcome":           "Welcome to %s",
		"title.license":           "License Agreement",
		"title.directory":         "Select Installation Directory",
		"title.options":           "Installation Options",
		"title.form":              "Configuration",
		"title.progress":          "Installing...",
		"title.summary":           "Installation Summary",
		"title.complete":          "Installation Complete",
		"title.info":              "Information",
		"desc.welcome":            "This will install %s on your computer.",
		"desc.welcome.version":    "This will install %s version %s on your computer.",
		"desc.directory":          "Choose the folder where you want to install the application.",
		"desc.progress.wait":      "Please wait while the installation completes.",
		"desc.summary.success":    "%s has been installed on your computer.",
		"footer.welcome":          "Click 'Continue' to proceed with the installation.",
		"label.install.dir":       "Installation Directory:",
		"label.required.space":    "Required space: ",
		"label.plan":              "Planned actions:",
		"label.errors":            "Errors:",
		"label.logfile":           "Log file: %s",
		"label.installed.to":      "Installed to:",
		"label.launch":            "Launch application after closing",
		"label.accept":            "I accept the terms of the license agreement",
		"label.admin":             "(admin)",
		"status.prepare":          "Preparing...",
		"status.failed":           "Installation Failed",
		"status.complete":         "Installation Complete",
		"status.no_tasks":         "No tasks to run",
		"msg.no_tasks":            "No tasks configured for installation.",
		"msg.ready":               "Review the details before installing.",
		"msg.content.file":        "Content would be loaded from: %s",
		"msg.license.file":        "License content would be loaded from: %s",
		"msg.license.empty":       "License content is empty.",
		"msg.scroll_end":          "Please scroll to the end of the license to continue.",
		"msg.accept_license":      "You must accept the license agreement to continue.",
		"msg.success":             "✓ The installation completed successfully!",
		"msg.failure":             "✗ The installation encountered errors.",
		"msg.field.required":      "%s is required",
		"msg.dir.required":        "Please select an installation directory.",
		"msg.dir.create":          "Cannot create installation directory: %v",
		"msg.dir.parent":          "Parent path is not a directory.",
		"msg.task.queue_fail":     "Failed to queue task: %v",
		"msg.task.start":          "Starting: %s",
		"msg.task.complete":       "Completed: %s",
		"msg.task.error":          "Error in %s: %v",
		"msg.install.failed":      "Installation failed: %v",
		"msg.install.in_progress": "Installation is still in progress",
		"footer.close":            "Click 'Close' to exit the installer.",
	},
	"zh": {
		"button.back":              "返回",
		"button.cancel":            "取消",
		"button.continue":          "继续",
		"button.install":           "安装",
		"button.finish":            "完成",
		"button.exit":              "退出",
		"button.close":             "关闭",
		"button.browse":            "浏览...",
		"dialog.cancel.title":      "取消安装",
		"dialog.cancel.msg":        "确定要取消安装吗？",
		"dialog.select.dir":        "选择目录",
		"dialog.select.file":       "选择文件",
		"dialog.validation.title":  "校验错误",
		"dialog.error.title":       "错误",
		"title.welcome":            "欢迎使用 %s",
		"title.license":            "许可协议",
		"title.directory":          "选择安装目录",
		"title.options":            "安装选项",
		"title.form":               "配置",
		"title.progress":           "正在安装...",
		"title.summary":            "安装摘要",
		"title.complete":           "安装完成",
		"title.info":               "信息",
		"desc.welcome":             "这将把 %s 安装到您的计算机上。",
		"desc.welcome.version":     "这将把 %s %s 版本安装到您的计算机上。",
		"desc.directory":           "请选择要安装应用的目录。",
		"desc.progress.wait":       "请稍候，正在完成安装。",
		"desc.summary.success":     "%s 已安装到您的计算机上。",
		"footer.welcome":           "点击“继续”开始安装。",
		"label.install.dir":        "安装目录：",
		"label.required.space":     "所需空间：",
		"label.plan":               "计划执行：",
		"label.errors":             "错误：",
		"label.logfile":            "日志文件：%s",
		"label.installed.to":       "安装位置：",
		"label.launch":             "关闭后启动应用",
		"label.accept":             "我已阅读并同意许可协议",
		"label.admin":              "（管理员）",
		"status.prepare":           "准备中...",
		"status.failed":            "安装失败",
		"status.complete":          "安装完成",
		"status.no_tasks":          "没有可执行的任务",
		"msg.no_tasks":             "未配置安装任务。",
		"msg.ready":                "安装前请先确认详细信息。",
		"msg.content.file":         "内容将从以下位置加载：%s",
		"msg.license.file":         "许可内容将从以下位置加载：%s",
		"msg.license.empty":        "许可内容为空。",
		"msg.scroll_end":           "请先滚动到协议末尾。",
		"msg.accept_license":       "请先勾选同意许可协议。",
		"msg.success":              "✓ 安装已成功完成！",
		"msg.failure":              "✗ 安装过程中出现错误。",
		"msg.field.required":       "%s 为必填项",
		"msg.dir.required":         "请选择安装目录。",
		"msg.dir.create":           "无法创建安装目录：%v",
		"msg.dir.parent":           "父路径不是目录。",
		"msg.task.queue_fail":      "任务入队失败：%v",
		"msg.task.start":           "开始：%s",
		"msg.task.complete":        "完成：%s",
		"msg.task.error":           "%s 出错：%v",
		"msg.install.failed":       "安装失败：%v",
		"msg.install.in_progress":  "安装仍在进行中",
		"footer.close":             "点击“关闭”退出安装程序。",
		"Demo Application":         "演示应用",
		"Welcome":                  "欢迎",
		"License Agreement":        "许可协议",
		"Installation Directory":   "安装目录",
		"Installation Options":     "安装选项",
		"Ready to Install":         "准备安装",
		"Installing":               "正在安装",
		"Complete":                 "完成",
		"Uninstall":                "卸载",
		"Confirm":                  "确认",
		"Uninstalling":             "正在卸载",
		"Create desktop shortcut":  "创建桌面快捷方式",
		"Add to PATH":              "添加到 PATH",
		"Start after installation": "安装后启动",
		`Welcome to Demo Application Installer!

This wizard will guide you through the installation of Demo Application.

Click Next to continue.`: `欢迎使用 Demo Application 安装向导！

本向导将引导您完成 Demo Application 的安装。

点击下一步继续。`,
		`This wizard will remove Demo Application from your system.

Click Next to continue.`: `本向导将从系统中移除 Demo Application。

点击下一步继续。`,
	},
}

const (
	i18nPrefix      = "@i18n:"
	i18nShortPrefix = "@t:"
)

func tr(ctx *core.InstallContext, key, fallback string) string {
	locale := resolveLocale(ctx)
	if val, ok := lookupTranslation(ctx, locale, key); ok {
		return val
	}
	if locale != "en" {
		if val, ok := lookupTranslation(ctx, "en", key); ok {
			return val
		}
	}
	return fallback
}

func trText(ctx *core.InstallContext, text string) string {
	if text == "" {
		return ""
	}

	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, i18nPrefix) || strings.HasPrefix(trimmed, i18nShortPrefix) {
		key := strings.TrimSpace(trimmed[strings.Index(trimmed, ":")+1:])
		if key == "" {
			return ""
		}
		return renderText(ctx, tr(ctx, key, key))
	}

	locale := resolveLocale(ctx)
	if translated, ok := lookupTranslation(ctx, locale, text); ok {
		return renderText(ctx, translated)
	}
	if trimmed != text {
		if translated, ok := lookupTranslation(ctx, locale, trimmed); ok {
			return renderText(ctx, translated)
		}
	}

	return renderText(ctx, text)
}

func lookupTranslation(ctx *core.InstallContext, locale, key string) (string, bool) {
	if ctx != nil {
		if value, ok := ctx.Get("i18n"); ok {
			if translated, ok := lookupTranslationValue(value, locale, key); ok {
				return translated, true
			}
		}
		if value, ok := ctx.Get("meta.i18n"); ok {
			if translated, ok := lookupTranslationValue(value, locale, key); ok {
				return translated, true
			}
		}
	}
	if bundle, ok := uiTranslations[locale]; ok {
		if val, ok := bundle[key]; ok {
			return val, true
		}
	}
	return "", false
}

func lookupTranslationValue(value any, locale, key string) (string, bool) {
	switch v := value.(type) {
	case map[string]map[string]string:
		if bundle, ok := v[locale]; ok {
			if val, ok := bundle[key]; ok {
				return val, true
			}
		}
	case map[string]map[string]any:
		if bundle, ok := v[locale]; ok {
			return lookupStringValue(bundle, key)
		}
	case map[string]string:
		if val, ok := v[key]; ok {
			return val, true
		}
	case map[string]any:
		if bundle, ok := v[locale]; ok {
			if translated, ok := lookupTranslationValue(bundle, locale, key); ok {
				return translated, true
			}
		}
		return lookupStringValue(v, key)
	}
	return "", false
}

func lookupStringValue(value any, key string) (string, bool) {
	switch v := value.(type) {
	case map[string]string:
		val, ok := v[key]
		return val, ok
	case map[string]any:
		if raw, ok := v[key]; ok {
			if str, ok := raw.(string); ok {
				return str, true
			}
		}
	}
	return "", false
}

func renderText(ctx *core.InstallContext, text string) string {
	if ctx == nil || text == "" {
		return text
	}
	return ctx.Render(text)
}

func resolveLocale(ctx *core.InstallContext) string {
	if ctx != nil {
		if v, ok := ctx.Get("meta.lang"); ok {
			return normalizeLocale(v)
		}
		if v, ok := ctx.Get("meta.locale"); ok {
			return normalizeLocale(v)
		}
	}
	if env := os.Getenv("LANG"); env != "" {
		return normalizeLocale(env)
	}
	return "en"
}

func normalizeLocale(value any) string {
	raw := strings.ToLower(strings.TrimSpace(strings.Split(fmt.Sprintf("%v", value), ".")[0]))
	if strings.HasPrefix(raw, "zh") {
		return "zh"
	}
	return "en"
}
