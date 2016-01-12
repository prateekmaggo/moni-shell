package main

import (
  "fmt"
  "os"
  "strings"
  "io/ioutil"
  "os/exec"
  "math/big"
  "time"
  "strconv"
)

func getContainerId() ([]string){

  var (
    lsOut []byte
    err error
  )
  // Append path with /monitor when runing inside Container
  liveContainerDirPath := "/var/run/docker/execdriver/native/"
  cmdLs := "ls"
  if lsOut, err = exec.Command(cmdLs, liveContainerDirPath).Output(); err != nil {
    fmt.Fprintln(os.Stderr, " Error listing live containers ")
    os.Exit(1)
  }
  liveContainerId := strings.Split(string(lsOut),"\n")
  return liveContainerId

}

func readCpuAcctUsage(dirPathCpuacct string) *big.Int {
  fileRead, err := ioutil.ReadFile(dirPathCpuacct+"cpuacct.usage")
  if err !=nil {
    panic(err)
  }
  bi := big.NewInt(0)
  bi.SetString(string(fileRead), 10)
  return bi
}

func readTotalSystemUsages() *big.Int {
  var (
    statData []string
    statCpu []string
    totalSystemCount = big.NewInt(0)
    convertIntToBig = big.NewInt(1)
  )
  fileRead, err := ioutil.ReadFile("/proc/stat")
  if err != nil {
    panic(err)
  }
  statData = strings.Split(string(fileRead),"\n")
  statCpu = strings.Split(string(statData[0]), "cpu")
  strCpuTrim := strings.TrimSpace(statCpu[1])
  strCpuValue := strings.Split(strCpuTrim, " ")
  totalSystemCount.SetInt64(0)
  for counter := 0 ; counter < len(strCpuValue) ; counter++  {
    if value, err := strconv.ParseInt(strCpuValue[counter], 10, 64); err == nil {
      convertIntToBig.SetInt64(value)
      totalSystemCount.Add(totalSystemCount, convertIntToBig)
    }
  }
  return totalSystemCount
}

func checkContainer(containerDirPath string) (bool) {
  if _, err := os.Stat(containerDirPath); err == nil {
    return true
  } else {
    return false
  }
}

func calculateContainerCpuUtilization(containerDirPath, containerID string) {
  var (
      preContainerUsages = big.NewInt(0)
      curContainerUsages = big.NewInt(1)
      preSystemCount = big.NewInt(3)
      curSystemCount = big.NewInt(4)
      systemCount = big.NewInt(5)
      containerCount = big.NewInt(6)
      finalCpuCount float64
      i = big.NewInt(8)
      containerExist bool
    )

    i.SetInt64(10000000)

    preSystemCount = readTotalSystemUsages()
    preSystemCount.Mul(preSystemCount, i)
    if containerExist = checkContainer(containerDirPath); containerExist {
      preContainerUsages = readCpuAcctUsage(containerDirPath)
    } else {
        containerExist = false
        finalCpuCount = 0.0
    }

    time.Sleep(1000 * time.Millisecond)

    curSystemCount = readTotalSystemUsages()
    curSystemCount.Mul(curSystemCount, i)
    if containerExist = checkContainer(containerDirPath); containerExist {
      curContainerUsages = readCpuAcctUsage(containerDirPath)
    } else {
      containerExist = false
      finalCpuCount = 0.0
    }

    systemCount.Sub(curSystemCount, preSystemCount)
    containerCount.Sub(curContainerUsages, preContainerUsages)

    x := containerCount.Uint64()
    y := systemCount.Uint64()
    finalCpuCount = float64(x)/float64(y)
    finalCpuCount = finalCpuCount * 100

    //return finalCpuCount, containerExist
    fmt.Println(finalCpuCount, containerID, containerExist)
}
 // Append path with /monitor when runing inside Container
func containerUtilization() {
  dirPathCpuacct := "/sys/fs/cgroup/cpuacct/"
  containerID := getContainerId()
  if len(containerID) == 1 {
    fmt.Println("Container doesn't exist")
    os.Exit(1)
  } else {

    for i := 0 ; i < len(containerID) - 1 ; i++ {
      containerDirPath := dirPathCpuacct + "docker/" + containerID[i] + "/"
      go calculateContainerCpuUtilization(containerDirPath, containerID[i])
    }
  }

}

func main() {
 for {
    containerUtilization()
    time.Sleep(30000 * time.Millisecond)
 }



}
