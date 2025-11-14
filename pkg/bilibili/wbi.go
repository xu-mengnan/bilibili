package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WBI密钥结构
type WBIKey struct {
	ImgKey string
	SubKey string
}

// 默认WBI密钥（降级方案）
var defaultWBIKey = WBIKey{
	ImgKey: "6536ef935693ef639889778317a124ab", // 默认图片密钥
	SubKey: "44aa19dd532868a0e7278589417478a8", // 默认子密钥
}

// NavResponse 用于解析获取WBI密钥的API响应
type NavResponse struct {
	Code int `json:"code"`
	Data struct {
		WbiImg struct {
			ImgUrl string `json:"img_url"`
			SubUrl string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

// GetWBIKey 获取WBI密钥
func GetWBIKey() WBIKey {
	// 创建HTTP客户端
	client := &http.Client{Timeout: 10 * time.Second}

	// 创建请求
	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		// 如果获取失败，使用默认密钥
		return defaultWBIKey
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		// 如果请求失败，使用默认密钥
		return defaultWBIKey
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 如果读取失败，使用默认密钥
		return defaultWBIKey
	}

	// 解析JSON
	var navResp NavResponse

	if err := json.Unmarshal(body, &navResp); err != nil || navResp.Code != 0 {
		// 如果解析失败或返回错误码，使用默认密钥
		return defaultWBIKey
	}

	// 提取密钥
	imgKey := extractKeyFromURL(navResp.Data.WbiImg.ImgUrl)
	subKey := extractKeyFromURL(navResp.Data.WbiImg.SubUrl)

	return WBIKey{
		ImgKey: imgKey,
		SubKey: subKey,
	}
}

// 从URL提取密钥
func extractKeyFromURL(urlStr string) string {
	// 从URL中提取文件名（不含扩展名）
	parts := strings.Split(urlStr, "/")
	filename := parts[len(parts)-1]
	return strings.Split(filename, ".")[0]
}

// mixinKeyEncTab 是WBI签名算法中使用的固定表
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

// getMixinKey 从原始密钥生成混合密钥
func getMixinKey(orig string) string {
	var str strings.Builder
	for _, v := range mixinKeyEncTab {
		if v < len(orig) {
			str.WriteByte(orig[v])
		}
	}
	return str.String()[:32]
}

// SignParams 对参数进行WBI签名
func SignParams(params url.Values, wbiKey WBIKey) url.Values {
	// 复制参数，避免修改原始参数
	signedParams := url.Values{}
	for k, v := range params {
		signedParams[k] = v
	}

	// 添加wts参数（当前时间戳）
	wts := strconv.FormatInt(time.Now().Unix(), 10)
	signedParams.Set("wts", wts)

	// 对参数按键名排序
	keys := make([]string, 0, len(signedParams))
	for k := range signedParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构造查询字符串，并过滤特殊字符
	var query strings.Builder
	for _, k := range keys {
		if query.Len() > 0 {
			query.WriteByte('&')
		}

		// 过滤掉特殊字符 !'()*
		value := signedParams.Get(k)
		value = strings.ReplaceAll(value, "!", "")
		value = strings.ReplaceAll(value, "'", "")
		value = strings.ReplaceAll(value, "(", "")
		value = strings.ReplaceAll(value, ")", "")
		value = strings.ReplaceAll(value, "*", "")

		query.WriteString(k)
		query.WriteByte('=')
		query.WriteString(value)
	}

	// 生成混合密钥
	mixinKey := getMixinKey(wbiKey.ImgKey + wbiKey.SubKey)

	// 计算w_rid
	hash := md5.Sum([]byte(query.String() + mixinKey))
	w_rid := hex.EncodeToString(hash[:])

	// 添加w_rid参数
	signedParams.Set("w_rid", w_rid)

	return signedParams
}
