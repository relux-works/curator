// Review-only: failed Muse shared capture preserves state and never observes a
// vendor, followed by recovery through the same public capture/status APIs.
func TestReviewMuseCaptureRefusalNoProbe(t *testing.T) {
 fx, req := captureFixture(t)
 fx.native[envregistry.Muse] = t.TempDir()
 auth := filepath.Join(fx.native[envregistry.Muse], "auth.json")
 payload := []byte("synthetic-muse\n")
 if err := os.WriteFile(auth,payload,0600); err != nil { t.Fatal(err) }
 provision := fx.request(envregistry.Muse)
 provision.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse:envregistry.IsolationShared}
 provision.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile:{envregistry.Muse:envregistry.CredentialStoreFile}}
 provision.Repair = true
 if _,err := Resolve(provision); err != nil { t.Fatal(err) }
 req.Resolve = provision; req.Resolve.Repair = false
 req.Resolve.Detect = func(envregistry.Adapter) string {t.Fatal("capture observed vendor");return "unknown"}
 req.ReloadMachine = func()(envregistry.MachineConfig,Policy,error){return provision.Machine,Policy{},nil}
 req.Capture = func(c *MaterializeCapture)([]byte,error){return []byte("fixture"),nil}
 if _,err := CaptureMaterialize(req); err != nil {t.Fatal(err)}
 if err := os.Remove(auth); err != nil {t.Fatal(err)}
 before,_ := captureFixtureState(t,fx)
 req.Capture = func(c *MaterializeCapture)([]byte,error){t.Fatal("absent native auth reached callback");return nil,nil}
 requireCaptureRefusal(t,req,DiagHomeStale)
 status,err := StatusOf(StatusRequest{Home:fx.home,Machine:req.Resolve.Machine,NativeHomeOf:req.Resolve.NativeHomeOf,OperatorXDG:fx.xdg,LaunchDir:fx.launch})
 if err != nil {t.Fatal(err)}
 row:=findHome(status,fx.profile,envregistry.Muse)
 if row==nil || row.Current {t.Fatalf("missing auth status: %+v",row)}
 after,_:=captureFixtureState(t,fx)
 if !bytes.Equal(before,after){t.Fatal("refusal changed source")}
 if err:=os.WriteFile(auth,payload,0600);err!=nil{t.Fatal(err)}
 req.Capture=func(c *MaterializeCapture)([]byte,error){return []byte("fixture"),nil}
 if _,err:=CaptureMaterialize(req);err!=nil{t.Fatal("restored native source",err)}
}
