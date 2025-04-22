package main

import (
	"fmt"

	"github.com/ziyi233/AIMetricFeeder/email"
)

func main() {
	cfg, err := email.LoadConfigFromFileOrEnv()
	if err != nil {
		fmt.Printf("加载邮件配置失败: %v\n", err)
		return
	}
	mail := email.Email{
		To:           []string{"1790060255@qq.com", "1193891280@qq.com"},
		Subject:      "数据库性能异常告警 - AIMetricFeeder",
		IsHTML:       true,
		TemplateName: "db_analysis.html",
		TemplateData: map[string]interface{}{
			"Title":              "数据库性能异常告警",
			"Subtitle":           "系统检测到数据库性能指标异常，AI已完成分析",
			"AISummary":          "发现3个慢查询和CPU使用率异常峰值，建议优化索引和查询语句",
			"RiskLevel":          "高风险",
			"RiskLevelLower":     "high",
			"AnalysisTime":       "2025-04-22 23:45:12",
			"AIDetails": `AI分析发现以下问题：

1. 慢查询问题：3个查询执行时间超过5秒，均来自用户管理模块
2. CPU使用率：数据库服务器CPU使用率在23:30达到95%，远高于正常水平
3. 索引缺失：user_logs表缺少针对operation_time字段的索引

问题SQL语句：
SELECT * FROM user_logs WHERE operation_time BETWEEN '2025-04-21' AND '2025-04-22' AND user_type = 'admin';

关键指标：
- CPU使用率：95%（基准值<70%）- 异常
- 平均查询响应时间：3.2秒（基准值<0.5秒）- 异常
- 活跃连接数：45（基准值<100）- 正常

优化建议：
1. 为user_logs表的operation_time字段创建索引
2. 优化查询语句，避免SELECT *，只查询必要字段
3. 考虑增加数据库服务器内存至16GB
4. 检查并终止异常耗时的后台任务`,
			"LogContent":         "2025-04-22 23:30:01 [WARNING] Slow query detected: 5.2s\n2025-04-22 23:30:05 [WARNING] CPU usage: 95%\n2025-04-22 23:31:12 [ERROR] Query timeout: SELECT * FROM user_logs WHERE...\n2025-04-22 23:32:45 [WARNING] Memory usage threshold exceeded: 85%\n2025-04-22 23:35:01 [INFO] Automatic query optimization attempted\n2025-04-22 23:40:22 [WARNING] Slow query detected: 6.1s\n2025-04-22 23:30:01 [WARNING] Slow query detected: 5.2s\n2025-04-22 23:30:05 [WARNING] CPU usage: 95%\n2025-04-22 23:31:12 [ERROR] Query timeout: SELECT * FROM user_logs WHERE...\n2025-04-22 23:32:45 [WARNING] Memory usage threshold exceeded: 85%\n2025-04-22 23:35:01 [INFO] Automatic query optimization attempted\n2025-04-22 23:40:22 [WARNING] Slow query detected: 6.1s\n2025-04-22 23:30:01 [WARNING] Slow query detected: 5.2s\n2025-04-22 23:30:05 [WARNING] CPU usage: 95%\n2025-04-22 23:31:12 [ERROR] Query timeout: SELECT * FROM user_logs WHERE...\n2025-04-22 23:32:45 [WARNING] Memory usage threshold exceeded: 85%\n2025-04-22 23:35:01 [INFO] Automatic query optimization attempted\n2025-04-22 23:40:22 [WARNING] Slow query detected: 6.1s\n2025-04-22 23:30:01 [WARNING] Slow query detected: 5.2s\n2025-04-22 23:30:05 [WARNING] CPU usage: 95%\n2025-04-22 23:31:12 [ERROR] Query timeout: SELECT * FROM user_logs WHERE...\n2025-04-22 23:32:45 [WARNING] Memory usage threshold exceeded: 85%\n2025-04-22 23:35:01 [INFO] Automatic query optimization attempted\n2025-04-22 23:40:22 [WARNING] Slow query detected: 6.1s",
			"Footer":             "AIMetricFeeder自动监控 | 报告ID: DB-20250422-001",
		},
	}
	err = email.NewMailer(cfg).Send(mail)
	if err != nil {
		fmt.Printf("邮件发送失败: %v\n", err)
	} else {
		fmt.Println("测试邮件发送成功，请查收收件箱！")
	}
}
