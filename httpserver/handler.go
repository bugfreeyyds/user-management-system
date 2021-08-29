package main

import (
    "net/http"
    "path"
    "strings"

    "github.com/gin-gonic/gin"

    "user-management-system/type/code"
    "user-management-system/httpserver/rpcclient"
    "user-management-system/utils"
)

// generate upload image file name
func generateImgName(fname, postfix string) string {
    ext := path.Ext(fname)
    fileName := strings.TrimSuffix(fname, ext)
    fileName = utils.Md5String(fileName + postfix)

    return fileName + ext
}

// login
func loginHandler(c *gin.Context) {
    // check params
    username := c.PostForm("username")
    passwd := c.PostForm("passwd")

    if len(passwd) != 32 {
        //logs2.Error("Invalid passwd:", passwd)
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeInvalidPasswd, "", nil))
        return
    }

    uuid := utils.GenerateToken(username)
    //logs2.Debug(uuid, " -- loginHandler access from:", username, "@", passwd)

    // communicate with rcp server
    ret, token, rsp := rpcclient.Login(map[string]string{"username":username, "passwd":passwd, "uuid":uuid})
    // set cookie
    if ret == http.StatusOK && token != "" {
        c.SetCookie("token", token, config.Logic.Tokenexpire, "/", config.Server.IP, false, true)
        //logs2.Debug(uuid, " -- Set token ", token, "with expire:", config.Logic.Tokenexpire)
    }

    //logs2.Debug(uuid, " -- Succ get response from backend with", rsp["code"], " and msg:", rsp["msg"])
    c.JSON(ret, rsp)
}

// logout
func logoutHandler(c* gin.Context) {
    // check params
    username := c.PostForm("username")
    token, err := c.Cookie("token")
    if err != nil {
        //logs2.Error("Failed to get token from cookie, err:", err.Error())
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeTokenNotFound, "", nil))
        return
    }

    if len(token) != 32 {
        //logs2.Error("Invalid token :", token)
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeInvalidToken, "", nil))
        return
    }
    uuid := utils.GenerateToken(username)
    //logs2.Debug(uuid, " -- logoutHandler access from:", username, " with token:", token)

    // communicate with rcp server
    ret, rsp := rpcclient.Logout(map[string]string{"username":username, "token":token, "uuid":uuid})

    //logs2.Debug(uuid, " -- Succ to get response from backend with ", rsp["code"], " and msg:", rsp["msg"])
    c.JSON(ret, rsp)
}

// edit nickname
func editNicknameHandler(c* gin.Context) {
    // check params
    username := c.PostForm("username")
    nickname := c.PostForm("newnickname")
    token, err := c.Cookie("token")
    //logs2.Debug("access from:", username, " with token:", token, " and newname:", nickname)
    if err != nil {
        //logs2.Error("Failed to get token from cookie, err:", err.Error())
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeTokenNotFound, "", nil))
        return
    }

    if len(token) != 32 {
        //logs2.Error("Invalid token :", token)
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeInvalidToken, "", nil))
        return
    }

    uuid := utils.GenerateToken(username)
    //logs2.Debug(uuid, " -- editNicknameHandler access from:", username, " with token:", token, " new nickname:", nickname)
    // communicate with rcp server
    ret, rsp := rpcclient.EditUserinfo(map[string]string{"username":username, "token":token, "nickname":nickname, "headurl":"", "mode": "1", "uuid":uuid})

    //logs2.Debug(uuid, " -- Succ to get response from backend with ", rsp["code"], " and msg:", rsp["msg"])
    c.JSON(ret, rsp)
}

// uploadHeadurlHandle
func uploadHeadurlHandler(c* gin.Context) {
    // check params
    username := c.Query("username")
    token, err := c.Cookie("token")
    //logs2.Debug("access from:", username, " with token:", token)
    if err != nil {
        //logs2.Error("Failed to get token from cookie, err:", err.Error())
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeTokenNotFound, "", nil))
        return
    }

    uuid := utils.GenerateToken(username)
    //logs2.Debug(uuid, " -- uploadHeadurlHandler access from:", username, " with token:", token)
    // step 1 : auth
    httpCode, tcpCode, msg := rpcclient.Auth(map[string]string{"username":username, "token":token, "uuid":uuid})
    if httpCode != http.StatusOK || tcpCode != 0 {
        //logs2.Error(uuid, " -- uploadHeadurlHandler Auth failed, msg:", msg)
        c.JSON(httpCode, rpcclient.FormatResponse(tcpCode, msg, nil))
        return
    }
    //logs2.Debug(uuid, " -- uploadHeadurlHandler Auth succ")
    // step 2 : save upload picture into file
    // save picture
    file, image, err := c.Request.FormFile("picture")
    if err != nil {
        //logs2.Error(uuid, " -- Failed to FormFile, err:", err.Error())
        c.JSON(http.StatusOK, rpcclient.FormatResponse(code.CodeFormFileFailed, "", nil))
    }

    //// check image
    if image == nil {
        //logs2.Error(uuid, " -- Failed to get image from formfile!")
        c.JSON(http.StatusOK, rpcclient.FormatResponse(code.CodeFormFileFailed, "", nil))
        return
    }
    //// check filesize
    size, err := utils.GetFileSize(file)
    if err != nil {
        //logs2.Error(uuid, " -- Failed to get filesize, err:", err.Error())
        c.JSON(http.StatusOK, rpcclient.FormatResponse(code.CodeFileSizeErr, "", nil))
        return
    }
    if size == 0 || size > config.Image.Maxsize * 1024 * 1024 {
        //logs2.Error(uuid, " -- Filesize illegal, size:", size)
        c.JSON(http.StatusOK, rpcclient.FormatResponse(code.CodeFileSizeErr, "", nil))
        return
    }
    //logs2.Debug(uuid, " -- uploadHeadurlHandler CheckImage succ")
    //// save
    imageName := generateImgName(image.Filename, username)
    fullPath  := config.Image.Savepath + imageName

    if err = c.SaveUploadedFile(image, fullPath); err != nil {
        //logs2.Error(uuid, " -- Failed to save file, err:", err.Error())
        c.JSON(http.StatusInternalServerError, rpcclient.FormatResponse(code.CodeInternalErr, "", nil))
        return
    }
    //logs2.Debug(uuid, " -- Succ to save upload image, path:", fullPath)

    // step 3 : update picture info
    imageURL := config.Image.Prefixurl + "/" + fullPath
    ret, editRsp := rpcclient.EditUserinfo(map[string]string{"username": username, "token": token, "nickname": "", "headurl": imageURL, "mode": "2", "uuid":uuid})
    //logs2.Debug(uuid, " -- editUserInfo response:", ret)
    c.JSON(ret, editRsp)
}

// get user info
func getUserinfoHandler(c* gin.Context) {
    // check params
    username := c.Query("username")
    token, err := c.Cookie("token")
    //logs2.Debug("access from:", username, " with token:", token)
    if err != nil {
        //logs2.Error("Failed to get token from cookie, err:", err.Error())
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeTokenNotFound, "", nil))
        return
    }

    if len(token) != 32 {
        //logs2.Error("Invalid token :", token)
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeInvalidToken, "", nil))
        return
    }

    uuid := utils.GenerateToken(username)
    //logs2.Debug(uuid, " -- getUserinfoHandler access from:", username, " with token:", token)
    // communicate with rcp server
    ret, rsp := rpcclient.GetUserinfo(map[string]string{"username":username, "token":token, "uuid":uuid})
    //logs2.Debug(uuid, " -- Succ to get response from backend with ", rsp["code"], " and msg:", rsp["msg"])
    c.JSON(ret, rsp)
}

