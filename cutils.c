#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "cutils.h"
#include <ctype.h>

/* This metod is delegated to init the fsize in input with the byte lenght of the given file */
void set_file_size(FILE *fp, int *fsize)
{  
  fseek(fp, 0, SEEK_END);
  *fsize = ftell(fp);
  rewind(fp);
  return;
}

long get_file_size(char *filename)
{  
  long fsize = 0;
  FILE *fp;

  fp = fopen(filename, "r");
  if(fp) {
    fseek(fp, 0, SEEK_END);
    fsize = ftell(fp);
    fclose(fp);
  }
  return fsize;
}


/* Read the file and return the content */
char *read_content(const char *filename)
{
  char *fcontent = NULL;
  int fsize = 0;
  FILE *fp;
  fp = fopen(filename, "r");
  if(fp) {
    set_file_size(fp, &fsize);
    printf("File size: %d bytes\n",fsize);
    fcontent = (char*) malloc(sizeof(char) * fsize);
    fread(fcontent, 1, fsize, fp);
    fclose(fp);
  }
  return fcontent;
}

/* Read the file and return the content */
void read_content_no_alloc(const char *filename, char *out)
{
  int fsize = 0;
  FILE *fp;
  fp = fopen(filename, "r");
  if(fp) {
    set_file_size(fp, &fsize);
    fread(out, 1, fsize, fp);
    fclose(fp);
  }
  return ;
}

// Lower the case of the string in input
void lowerize_string(char *data){
  int i = 0; // Don't break c99 :/i
  for(; data[i]; i++){
    data[i] = tolower(data[i]);
  } 
}


int verify_presence_data_insensitive(char *data, const char *to_find)
{
  /* Manage None input */
  if(data != NULL && to_find != NULL){

    // lowering the line
    lowerize_string(data);
    /* Finding string ... */
    if(strstr(data,to_find)!=NULL)
    {
      // printf("String %s found ..\n",to_find);
      return 1;
    }
    //printf("String %s NOT found ..\n",to_find);
    return 0; 
  }
  else
    // Error - input
    return -1;
}


int verify_presence_data(const char *data, const char *to_find)
{
  /* Manage None input */
  if(data != NULL && to_find != NULL){
    /* Finding string ... */
    if(strstr(data,to_find)!=NULL)
    {
      // printf("String %s found ..\n",to_find);
      return 1;
    }
    //printf("String %s NOT found ..\n",to_find);
    return 0; 
  }
  else
    // Error - input
    return -1;
}

int verify_presence_filename(const char *filename, const char *to_find)
{

  /* Manage None input */
  if(filename != NULL && to_find != NULL){

    const char *content;
    content =(char *) read_content(filename);

    /* Finding string ... */
    return verify_presence_data(content,to_find);
  }

  printf("Seems that the filename or the string to find is NULL :/ ..."); 
  return -1;
}

/*int main(int argc, char **argv)
  {
  char *content;
  printf("%ld",get_file_size("/tmp/test1"));
  content = read_content("filename.txt");
  verify_presence_data(content,"provaT");
  verify_presence_filename(content,"filename.txt");
  free(content);
  return 0;
  }*/   


