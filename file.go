package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/kataras/iris/context"

	"golang.org/x/net/context/ctxhttp"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"

	"github.com/gw123/glog"
	"go.uber.org/zap"

	"github.com/kataras/iris"

	stdCtx "context"
)

const (
	FaasEnvProd     = "prod"
	FaasEnvEnOntest = "ontest"
)

func main() {
	app := iris.New()
	app.Any("/", func(ctx context.Context) {
		file_url := ctx.URLParam("file_url")
		funcname := ctx.URLParam("func_name")
		userId := ctx.URLParam("user_id")
		token := ctx.URLParam("group_user_token")
		runEnv := ctx.URLParam("runenv")
		refer := ctx.GetHeader("Referer")

		faasEnv := FaasEnvProd
		if strings.Contains(refer, "localhost") || strings.Contains(refer, "ontest") {
			faasEnv = FaasEnvEnOntest
		}
		glog.Infof("file_url %s, funanem: %s , token: %s, runEnb: %s", file_url, funcname, token, runEnv)
		if file_url == "" {
			ctx.WriteString("您可能使用旧版本函数，建议尝试在faasui页面上尝试发布代码后尝试重新点击 VSCODE 按钮")
			return
		}

		if token == "" {
			ctx.WriteString("参数不合法")
			return
		}

		glog.Infof("token %s ,user_id %s, refer : %s", token, userId, refer)
		//io.Copy(ctx.ResponseWriter(), ctx.Request().Body)
		distDir, err := Fetch(ctx.Request().Context(), &FetchPkg{
			Token:        token,
			FileUrl:      file_url,
			FunctionName: funcname,
			UserID:       userId,
			RunEnv:       runEnv,
			FaasEnv:      faasEnv,
		}, true)

		if err != nil {
			ctx.WriteString(err.Error())
			return
		}

		ctx.Request().Body.Close()
		// /#/
		//ctx.Redirect("http://172.21.206.63:8080/?folder=" + distDir)
		if faasEnv == FaasEnvProd {
			ctx.Redirect("http://172.21.125.26:8080/#" + distDir)
		} else {
			ctx.Redirect("http://172.21.206.63:8080/#" + distDir)
		}
	})

	app.Any("/download", func(ctx context.Context) {
		file_url := ctx.URLParam("file_url")
		funcname := ctx.URLParam("func_name")
		userId := ctx.URLParam("user_id")
		token := ctx.URLParam("group_user_token")
		runEnv := ctx.URLParam("runenv")
		refer := ctx.GetHeader("Referer")

		faasEnv := FaasEnvProd
		if strings.Contains(refer, "localhost") || strings.Contains(refer, "ontest") {
			faasEnv = FaasEnvEnOntest
		}
		glog.Infof("file_url %s, funanem: %s , token: %s, runEnb: %s", file_url, funcname, token, runEnv)
		if file_url == "" {
			ctx.WriteString("您可能使用旧版本函数，建议尝试在faasui页面上尝试发布代码后尝试重新点击 VSCODE 按钮")
			return
		}

		if token == "" {
			ctx.WriteString("参数不合法")
			return
		}

		glog.Infof("token %s ,user_id %s, refer : %s", token, userId, refer)
		//io.Copy(ctx.ResponseWriter(), ctx.Request().Body)
		distDir, err := Fetch(ctx.Request().Context(), &FetchPkg{
			Token:        token,
			FileUrl:      file_url,
			FunctionName: funcname,
			UserID:       userId,
			RunEnv:       runEnv,
			FaasEnv:      faasEnv,
		}, true)

		if err != nil {
			ctx.WriteString(err.Error())
			return
		}

		tmpFile := funcname + ".download"
		tmpPath := filepath.Join(os.TempDir(), tmpFile)
		Zip(distDir, tmpPath)
		ctx.SendFile(tmpPath, funcname+".zip")
	})

	app.Any("/choose", func(ctx context.Context) {
		uri := ctx.Request().RequestURI
		uri = strings.ReplaceAll(uri, "/choose", "/")
		strTpl := `
<head> <title>工作区已经存在是否覆盖当前编辑器代码</title> </head>
<body>
<script>
function myFunction(){
	var r=confirm("工作区已经存在是否覆盖当前编辑器代码!");
	if (r==true){
      window.location.href = "%s&"
	}else{
      window.location.href = "%s&"
	}
}
</script>
</body>
`
		ctx.WriteString(strTpl)
	})

	app.Run(iris.Addr(":8081"), iris.WithoutBodyConsumptionOnUnmarshal, iris.WithOptimizations)
}

type FetchPkg struct {
	Token        string `json:"group_user_token"`
	UserID       string `json:"user_id"`
	FileUrl      string `json:"file_url"`
	FunctionName string `json:"function_name"`
	RunEnv       string `json:"run_env"`
	FaasEnv      string `json:"faas_env"`
}

func Fetch(ctx stdCtx.Context, fpkg *FetchPkg, isRemove bool) (string, error) {
	logger := glog.DefaultLogger().WithField("svc", "fetch")
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logger.Info("fetch request done", zap.Duration("elapsed_time", elapsed))
	}()

	sharedVolumePath := os.Getenv("SharedVolumePath")
	if sharedVolumePath == "" {
		sharedVolumePath = "/home/coder/funcs"
	}

	filename := fpkg.FunctionName + "-" + fpkg.UserID
	distDir := filepath.Join(sharedVolumePath, filename)
	if _, err := os.Stat(distDir); err == nil {
		if isRemove {
			err := os.RemoveAll(distDir)
			if err != nil {
				glog.WithErr(err).Errorf("os.remove file err: %s", distDir)
			}
		} else {
			logger.Info("目录已经存在 - 跳过下载代码", zap.String("requested_file", fpkg.FileUrl), zap.String("shared_volume_path", distDir))
			return distDir, err
		}
	}

	tmpFile := filename + ".tmp"
	tmpPath := filepath.Join(os.TempDir(), tmpFile)
	err := DownloadUrl(ctx, http.DefaultClient, fpkg.FileUrl, tmpPath)
	if err != nil {
		e := "failed to download url"
		logger.Error(e, zap.Error(err), zap.String("url", fpkg.FileUrl))
		return "", err
	}
	glog.Infof("download to tmp file %s", tmpPath)

	err = unarchive(tmpPath, distDir)
	if err != nil {
		logger.WithError(err).Error("error unarchive file", zap.String("original_path", tmpPath), zap.String("distDir", distDir))
		return "", err
	}
	os.Remove(tmpPath)

	err = os.MkdirAll(distDir+"/.faas", 0766)
	if err != nil {
		logger.WithError(err).Error("os mkdir file", zap.String("distDir", distDir+".faas"))
		return "", err
	}

	confTpl := "server: http://bcs-faas-ui-ontest.chj.cloud\n" +
		"group-user-token: " + fpkg.Token + "\n"

	if fpkg.FaasEnv == FaasEnvProd {
		confTpl = "server: http://bcs-faas-ui.chj.cloud\n" +
			"group-user-token: " + fpkg.Token + "\n"
	}

	glog.Infof("configTpl : %+v", confTpl)

	file, err := os.OpenFile(distDir+"/.faas/config", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		glog.WithErr(err).Errorf("打开.faas/config文件失败，请查看是否有权限")
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(confTpl)
	if err != nil {
		glog.WithErr(err).Errorf("写入.faas/config文件失败，请查看是否有权限")
		return "", err
	}

	funcTpl := `
functions:
  %s:
    executor_type: poolmgr
    run_env: %s 
    entrypoint: Handler
    source: .`

	funcConf := fmt.Sprintf(funcTpl, fpkg.FunctionName, fpkg.RunEnv)
	file, err = os.OpenFile(distDir+"/.faas/functions.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		glog.WithErr(err).Errorf("打开.faas/functions.yaml文件失败，请查看是否有权限")
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(funcConf)
	if err != nil {
		glog.WithErr(err).Errorf("写入.faas/functions.yaml文件失败，请查看是否有权限")
		return "", err
	}

	logger.Info("successfully placed", zap.String("location", distDir), zap.String("location", distDir+"/.faas/config"))
	return distDir, nil
}

func rename(src string, dst string) error {
	err := os.Rename(src, dst)
	if err != nil {
		return errors.Wrap(err, "failed to move file")
	}
	return nil
}

// unarchive is a function that unzips a zip file to destination
func unarchive(src string, dst string) error {
	err := archiver.Zip.Open(src, dst)
	if err != nil {
		return errors.Wrap(err, "failed to unzip file")
	}
	return nil
}

// 去掉最外面的目录
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	firstDir := ""

	var ignoreKeywords []string
	ignore := viper.GetString("ignore")
	ignore = strings.TrimSpace(ignore)
	if strings.Contains(ignore, ",") {
		ignoreKeywords = strings.Split(ignore, ",")
	}

	ignoreKeywords = append(ignoreKeywords, ".git/", ".idea/", ".ssh/", "node_modules/", ".log", ".zip", ".gz", "vendor/", "dist")
	fmt.Printf("忽略路径 %+v , 可以在配置文件中配置 ignore 添加过滤路径\n", ignoreKeywords)
	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		for _, ignoreKeyword := range ignoreKeywords {
			if strings.TrimSuffix(ignoreKeyword, "/") == path {
				if viper.GetBool("debug") {
					fmt.Printf("ignore %80s \t --contains-- \t%s\n", path, ignoreKeyword)
				}
				return nil
			}

			if strings.Contains(path, ignoreKeyword) {
				if viper.GetBool("debug") {
					fmt.Printf("ignore %80s \t --contains-- \t%s\n", path, ignoreKeyword)
				}
				return nil
			}
		}

		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 去掉最外面的目录
		// header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		header.Name = strings.TrimPrefix(path, srcFile+"/")
		if info.IsDir() {
			header.Name += "/"
			if firstDir == "" {
				firstDir = header.Name
				return nil
			}
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func DownloadUrl(ctx stdCtx.Context, httpClient *http.Client, url string, localPath string) error {
	resp, err := ctxhttp.Get(ctx, httpClient, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	w, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	// flushing write buffer to file
	err = w.Sync()
	if err != nil {
		return err
	}

	err = os.Chmod(localPath, 0600)
	if err != nil {
		return err
	}

	return nil
}
