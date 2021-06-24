namespace Module {
    using System;
    using System.IO;
    
    public static class CSTest {
        public static void Main(string[] args) {
            string path = "C:\\Users\\meet\\Desktop\\CSBEACTEST.txt";
            using (FileStream fs = File.Create(path))
            {
                System.IO.File.WriteAllText(path, "Text to add to the file\n");
            }
        } 
    }
}